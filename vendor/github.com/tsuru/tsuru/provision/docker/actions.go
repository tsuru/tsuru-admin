// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	"github.com/tsuru/config"
	"github.com/tsuru/tsuru/action"
	"github.com/tsuru/tsuru/app/image"
	"github.com/tsuru/tsuru/event"
	"github.com/tsuru/tsuru/log"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/docker/container"
	"github.com/tsuru/tsuru/router"
	"gopkg.in/mgo.v2/bson"
)

type runContainerActionsArgs struct {
	app              provision.App
	processName      string
	imageID          string
	commands         []string
	destinationHosts []string
	writer           io.Writer
	isDeploy         bool
	buildingImage    string
	provisioner      *dockerProvisioner
	exposedPort      string
	event            *event.Event
}

type containersToAdd struct {
	Quantity int
	Status   provision.Status
}

type changeUnitsPipelineArgs struct {
	app         provision.App
	writer      io.Writer
	toAdd       map[string]*containersToAdd
	toRemove    []container.Container
	toHost      string
	imageId     string
	provisioner *dockerProvisioner
	appDestroy  bool
	exposedPort string
	event       *event.Event
}

type callbackFunc func(*container.Container, chan *container.Container) error

type rollbackFunc func(*container.Container)

func runInContainers(containers []container.Container, callback callbackFunc, rollback rollbackFunc, parallel bool) error {
	if len(containers) == 0 {
		return nil
	}
	workers, _ := config.GetInt("docker:max-workers")
	if workers == 0 {
		workers = len(containers)
	}
	step := len(containers) / workers
	if len(containers)%workers != 0 {
		step += 1
	}
	toRollback := make(chan *container.Container, len(containers))
	errs := make(chan error, len(containers))
	var wg sync.WaitGroup
	runFunc := func(start, end int) error {
		defer wg.Done()
		for i := start; i < end; i++ {
			err := callback(&containers[i], toRollback)
			if err != nil {
				errs <- err
				return err
			}
		}
		return nil
	}
	for i := 0; i < len(containers); i += step {
		end := i + step
		if end > len(containers) {
			end = len(containers)
		}
		wg.Add(1)
		if parallel {
			go runFunc(i, end)
		} else {
			err := runFunc(i, end)
			if err != nil {
				break
			}
		}
	}
	wg.Wait()
	close(errs)
	close(toRollback)
	if err := <-errs; err != nil {
		if rollback != nil {
			for c := range toRollback {
				rollback(c)
			}
		}
		return err
	}
	return nil
}

func checkCanceled(evt *event.Event) error {
	if evt == nil {
		return nil
	}
	canceled, err := evt.AckCancel()
	if err != nil {
		log.Errorf("unable to check if event should be canceled, ignoring: %s", err)
		return nil
	}
	if canceled {
		return ErrDeployCanceled
	}
	return nil
}

var insertEmptyContainerInDB = action.Action{
	Name: "insert-empty-container",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		initialStatus := provision.StatusCreated
		if args.isDeploy {
			initialStatus = provision.StatusBuilding
		}
		contName := args.app.GetName() + "-" + randomString()
		cont := container.Container{
			AppName:       args.app.GetName(),
			ProcessName:   args.processName,
			Type:          args.app.GetPlatform(),
			Name:          contName,
			Status:        initialStatus.String(),
			Image:         args.imageID,
			BuildingImage: args.buildingImage,
			ExposedPort:   args.exposedPort,
		}
		coll := args.provisioner.Collection()
		defer coll.Close()
		if err := coll.Insert(cont); err != nil {
			log.Errorf("error on inserting container into database %s - %s", cont.Name, err)
			return nil, err
		}
		return cont, nil
	},
	Backward: func(ctx action.BWContext) {
		c := ctx.FWResult.(container.Container)
		args := ctx.Params[0].(runContainerActionsArgs)
		coll := args.provisioner.Collection()
		defer coll.Close()
		coll.Remove(bson.M{"name": c.Name})
	},
}

var updateContainerInDB = action.Action{
	Name: "update-database-container",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		coll := args.provisioner.Collection()
		defer coll.Close()
		cont := ctx.Previous.(container.Container)
		err := coll.Update(bson.M{"name": cont.Name}, cont)
		if err != nil {
			log.Errorf("error on updating container into database %s - %s", cont.ID, err)
			return nil, err
		}
		return cont, nil
	},
	Backward: func(ctx action.BWContext) {
	},
}

var createContainer = action.Action{
	Name: "create-container",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		cont := ctx.Previous.(container.Container)
		log.Debugf("create container for app %s, based on image %s, with cmds %s", args.app.GetName(), args.imageID, args.commands)
		var building bool
		if args.buildingImage != "" {
			building = true
		}
		err := cont.Create(&container.CreateArgs{
			ImageID:          args.imageID,
			Commands:         args.commands,
			App:              args.app,
			Deploy:           args.isDeploy,
			Provisioner:      args.provisioner,
			DestinationHosts: args.destinationHosts,
			ProcessName:      args.processName,
			Building:         building,
		})
		if err != nil {
			log.Errorf("error on create container for app %s - %s", args.app.GetName(), err)
			return nil, err
		}
		return cont, nil
	},
	Backward: func(ctx action.BWContext) {
		c := ctx.FWResult.(container.Container)
		args := ctx.Params[0].(runContainerActionsArgs)
		err := args.provisioner.Cluster().RemoveContainer(docker.RemoveContainerOptions{ID: c.ID})
		if err != nil {
			log.Errorf("Failed to remove the container %q: %s", c.ID, err)
		}
	},
}

var setContainerID = action.Action{
	Name: "set-container-id",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		coll := args.provisioner.Collection()
		defer coll.Close()
		cont := ctx.Previous.(container.Container)
		err := coll.Update(bson.M{"name": cont.Name}, bson.M{"$set": bson.M{"id": cont.ID}})
		if err != nil {
			log.Errorf("error on setting container ID %s - %s", cont.Name, err)
			return nil, err
		}
		return cont, nil
	},
	Backward: func(ctx action.BWContext) {
	},
}

var stopContainer = action.Action{
	Name: "stop-container",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		cont := ctx.Previous.(container.Container)
		err := cont.SetStatus(args.provisioner, provision.StatusStopped, false)
		if err != nil {
			return nil, err
		}
		return cont, nil
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(runContainerActionsArgs)
		c := ctx.FWResult.(container.Container)
		c.SetStatus(args.provisioner, provision.StatusCreated, false)
	},
}

var setNetworkInfo = action.Action{
	Name: "set-network-info",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		c := ctx.Previous.(container.Container)
		args := ctx.Params[0].(runContainerActionsArgs)
		info, err := c.NetworkInfo(args.provisioner)
		if err != nil {
			return nil, err
		}
		c.IP = info.IP
		c.HostPort = info.HTTPHostPort
		return c, nil
	},
}

var startContainer = action.Action{
	Name: "start-container",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		c := ctx.Previous.(container.Container)
		log.Debugf("starting container %s", c.ID)
		err := c.Start(&container.StartArgs{
			Provisioner: args.provisioner,
			App:         args.app,
			Deploy:      args.isDeploy,
		})
		if err != nil {
			log.Errorf("error on start container %s - %s", c.ID, err)
			return nil, err
		}
		return c, nil
	},
	Backward: func(ctx action.BWContext) {
		c := ctx.FWResult.(container.Container)
		args := ctx.Params[0].(runContainerActionsArgs)
		err := args.provisioner.Cluster().StopContainer(c.ID, 10)
		if err != nil {
			log.Errorf("Failed to stop the container %q: %s", c.ID, err)
		}
	},
}

var rollbackNotice = func(ctx action.FWContext, err error) {
	args := ctx.Params[0].(changeUnitsPipelineArgs)
	if args.writer != nil {
		fmt.Fprintf(args.writer, "\n**** ROLLING BACK AFTER FAILURE ****\n ---> %s <---\n", err)
	}
}

var provisionAddUnitsToHost = action.Action{
	Name: "provision-add-units-to-host",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		containers, err := addContainersWithHost(&args)
		if err != nil {
			return nil, err
		}
		return containers, nil
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		containers := ctx.FWResult.([]container.Container)
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		units := len(containers)
		fmt.Fprintf(w, "\n---- Destroying %d created %s ----\n", units, pluralize("unit", units))
		runInContainers(containers, func(cont *container.Container, _ chan *container.Container) error {
			err := cont.Remove(args.provisioner)
			if err != nil {
				log.Errorf("Error removing added container %s: %s", cont.ID, err)
				return nil
			}
			fmt.Fprintf(w, " ---> Destroyed unit %s [%s]\n", cont.ShortID(), cont.ProcessName)
			return nil
		}, nil, true)
	},
	OnError:   rollbackNotice,
	MinParams: 1,
}

var bindAndHealthcheck = action.Action{
	Name: "bind-and-healthcheck",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		webProcessName, err := image.GetImageWebProcessName(args.imageId)
		if err != nil {
			log.Errorf("[WARNING] cannot get the name of the web process: %s", err)
		}
		newContainers := ctx.Previous.([]container.Container)
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		doHealthcheck := true
		for _, c := range args.toRemove {
			if c.Status == provision.StatusError.String() || c.Status == provision.StatusStopped.String() {
				doHealthcheck = false
				break
			}
		}
		fmt.Fprintf(writer, "\n---- Binding and checking %d new %s ----\n", len(newContainers), pluralize("unit", len(newContainers)))
		return newContainers, runInContainers(newContainers, func(c *container.Container, toRollback chan *container.Container) error {
			unit := c.AsUnit(args.app)
			err := args.app.BindUnit(&unit)
			if err != nil {
				return err
			}
			toRollback <- c
			if doHealthcheck && c.ProcessName == webProcessName {
				err = runHealthcheck(c, writer)
				if err != nil {
					return err
				}
			}
			err = args.provisioner.runRestartAfterHooks(c, writer)
			if err != nil {
				return err
			}
			fmt.Fprintf(writer, " ---> Bound and checked unit %s [%s]\n", c.ShortID(), c.ProcessName)
			return nil
		}, func(c *container.Container) {
			unit := c.AsUnit(args.app)
			err := args.app.UnbindUnit(&unit)
			if err != nil {
				log.Errorf("Unable to unbind unit %q: %s", c.ID, err)
			}
		}, true)
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		newContainers := ctx.FWResult.([]container.Container)
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		units := len(newContainers)
		fmt.Fprintf(w, "\n---- Unbinding %d created %s ----\n", units, pluralize("unit", units))
		runInContainers(newContainers, func(c *container.Container, _ chan *container.Container) error {
			unit := c.AsUnit(args.app)
			err := args.app.UnbindUnit(&unit)
			if err != nil {
				log.Errorf("Removed binding for unit %q: %s", c.ID, err)
				return nil
			}
			fmt.Fprintf(w, " ---> Removed bind for unit %s [%s]\n", c.ShortID(), c.ProcessName)
			return nil
		}, nil, true)
	},
	OnError: rollbackNotice,
}

var addNewRoutes = action.Action{
	Name: "add-new-routes",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		webProcessName, err := image.GetImageWebProcessName(args.imageId)
		if err != nil {
			log.Errorf("[WARNING] cannot get the name of the web process: %s", err)
		}
		newContainers := ctx.Previous.([]container.Container)
		r, err := getRouterForApp(args.app)
		if err != nil {
			return nil, err
		}
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		if len(newContainers) > 0 {
			fmt.Fprintf(writer, "\n---- Adding routes to new units ----\n")
		}
		var routesToAdd []*url.URL
		for i, c := range newContainers {
			if c.ProcessName != webProcessName {
				continue
			}
			if c.ValidAddr() {
				routesToAdd = append(routesToAdd, c.Address())
				newContainers[i].Routable = true
			}
		}
		if len(routesToAdd) == 0 {
			return newContainers, nil
		}
		err = r.AddRoutes(args.app.GetName(), routesToAdd)
		if err != nil {
			r.RemoveRoutes(args.app.GetName(), routesToAdd)
			return nil, err
		}
		for _, c := range newContainers {
			if c.Routable {
				fmt.Fprintf(writer, " ---> Added route to unit %s [%s]\n", c.ShortID(), c.ProcessName)
			}
		}
		return newContainers, nil
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		newContainers := ctx.FWResult.([]container.Container)
		r, err := getRouterForApp(args.app)
		if err != nil {
			log.Errorf("[add-new-routes:Backward] Error geting router: %s", err)
		}
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		fmt.Fprintf(w, "\n---- Removing routes from created units ----\n")
		var routesToRemove []*url.URL
		for _, c := range newContainers {
			if c.Routable {
				routesToRemove = append(routesToRemove, c.Address())
			}
		}
		if len(routesToRemove) == 0 {
			return
		}
		err = r.RemoveRoutes(args.app.GetName(), routesToRemove)
		if err != nil {
			log.Errorf("[add-new-routes:Backward] Error removing route for [%v]: %s", routesToRemove, err)
			return
		}
		for _, c := range newContainers {
			if c.Routable {
				fmt.Fprintf(w, " ---> Removed route from unit %s [%s]\n", c.ShortID(), c.ProcessName)
			}
		}
	},
	OnError: rollbackNotice,
}

var setRouterHealthcheck = action.Action{
	Name:    "set-router-healthcheck",
	OnError: rollbackNotice,
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		newContainers := ctx.Previous.([]container.Container)
		r, err := getRouterForApp(args.app)
		if err != nil {
			return nil, err
		}
		hcRouter, ok := r.(router.CustomHealthcheckRouter)
		if !ok {
			return newContainers, nil
		}
		yamlData, err := image.GetImageTsuruYamlData(args.imageId)
		if err != nil {
			return nil, err
		}
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		hcData := yamlData.Healthcheck.ToRouterHC()
		msg := fmt.Sprintf("Path: %s", hcData.Path)
		if hcData.Status != 0 {
			msg = fmt.Sprintf("%s, Status: %d", msg, hcData.Status)
		}
		if hcData.Body != "" {
			msg = fmt.Sprintf("%s, Body: %s", msg, hcData.Body)
		}
		fmt.Fprintf(writer, "\n---- Setting router healthcheck (%s) ----\n", msg)
		err = hcRouter.SetHealthcheck(args.app.GetName(), hcData)
		return newContainers, err
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		r, err := getRouterForApp(args.app)
		if err != nil {
			log.Errorf("[set-router-healthcheck:Backward] Error getting router: %s", err)
			return
		}
		hcRouter, ok := r.(router.CustomHealthcheckRouter)
		if !ok {
			return
		}
		currentImageName, _ := image.AppCurrentImageName(args.app.GetName())
		yamlData, err := image.GetImageTsuruYamlData(currentImageName)
		if err != nil {
			log.Errorf("[set-router-healthcheck:Backward] Error getting yaml data: %s", err)
		}
		hcData := yamlData.Healthcheck.ToRouterHC()
		err = hcRouter.SetHealthcheck(args.app.GetName(), hcData)
		if err != nil {
			log.Errorf("[set-router-healthcheck:Backward] Error setting healthcheck: %s", err)
		}
	},
}

var removeOldRoutes = action.Action{
	Name: "remove-old-routes",
	Forward: func(ctx action.FWContext) (result action.Result, err error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		if err = checkCanceled(args.event); err != nil {
			return nil, err
		}
		result = ctx.Previous
		if args.appDestroy {
			defer func() {
				if err != nil {
					log.Errorf("ignored error during remove routes in app destroy: %s", err)
				}
				err = nil
			}()
		}
		r, err := getRouterForApp(args.app)
		if err != nil {
			return
		}
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		if len(args.toRemove) > 0 {
			fmt.Fprintf(writer, "\n---- Removing routes from old units ----\n")
		}
		currentImageName, err := image.AppCurrentImageName(args.app.GetName())
		if err != nil && err != image.ErrNoImagesAvailable {
			return
		}
		webProcessName, err := image.GetImageWebProcessName(currentImageName)
		if err != nil {
			log.Errorf("[WARNING] cannot get the name of the web process for route removal: %s", err)
		}
		var routesToRemove []*url.URL
		for i, c := range args.toRemove {
			if c.ProcessName != webProcessName {
				continue
			}
			if c.ValidAddr() {
				routesToRemove = append(routesToRemove, c.Address())
				args.toRemove[i].Routable = true
			}
		}
		if len(routesToRemove) == 0 {
			return
		}
		err = r.RemoveRoutes(args.app.GetName(), routesToRemove)
		if err != nil {
			if !args.appDestroy {
				r.AddRoutes(args.app.GetName(), routesToRemove)
			}
			return
		}
		for _, c := range args.toRemove {
			if c.Routable {
				fmt.Fprintf(writer, " ---> Removed route from unit %s [%s]\n", c.ShortID(), c.ProcessName)
			}
		}
		return
	},
	Backward: func(ctx action.BWContext) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		r, err := getRouterForApp(args.app)
		if err != nil {
			log.Errorf("[remove-old-routes:Backward] Error geting router: %s", err)
		}
		w := args.writer
		if w == nil {
			w = ioutil.Discard
		}
		fmt.Fprintf(w, "\n---- Adding back routes to old units ----\n")
		var routesToAdd []*url.URL
		for _, c := range args.toRemove {
			if c.Routable {
				routesToAdd = append(routesToAdd, c.Address())
			}
		}
		if len(routesToAdd) == 0 {
			return
		}
		err = r.AddRoutes(args.app.GetName(), routesToAdd)
		if err != nil {
			log.Errorf("[remove-old-routes:Backward] Error adding back route for [%v]: %s", routesToAdd, err)
			return
		}
		for _, c := range args.toRemove {
			if c.Routable {
				fmt.Fprintf(w, " ---> Added route to unit %s [%s]\n", c.ShortID(), c.ProcessName)
			}
		}
	},
	OnError:   rollbackNotice,
	MinParams: 1,
}

var provisionRemoveOldUnits = action.Action{
	Name: "provision-remove-old-units",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		total := len(args.toRemove)
		fmt.Fprintf(writer, "\n---- Removing %d old %s ----\n", total, pluralize("unit", total))
		runInContainers(args.toRemove, func(c *container.Container, toRollback chan *container.Container) error {
			err := c.Remove(args.provisioner)
			if err != nil {
				log.Errorf("Ignored error trying to remove old container %q: %s", c.ID, err)
			}
			fmt.Fprintf(writer, " ---> Removed old unit %s [%s]\n", c.ShortID(), c.ProcessName)
			return nil
		}, nil, true)
		return ctx.Previous, nil
	},
	Backward: func(ctx action.BWContext) {
	},
	OnError:   rollbackNotice,
	MinParams: 1,
}

var provisionUnbindOldUnits = action.Action{
	Name: "provision-unbind-old-units",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		writer := args.writer
		if writer == nil {
			writer = ioutil.Discard
		}
		total := len(args.toRemove)
		fmt.Fprintf(writer, "\n---- Unbinding %d old %s ----\n", total, pluralize("unit", total))
		runInContainers(args.toRemove, func(c *container.Container, toRollback chan *container.Container) error {
			unit := c.AsUnit(args.app)
			err := args.app.UnbindUnit(&unit)
			if err != nil {
				log.Errorf("Ignored error trying to unbind old container %q: %s", c.ID, err)
			}
			fmt.Fprintf(writer, " ---> Removed bind for old unit %s [%s]\n", c.ShortID(), c.ProcessName)
			return nil
		}, nil, true)
		return ctx.Previous, nil
	}, Backward: func(ctx action.BWContext) {
	},
	OnError:   rollbackNotice,
	MinParams: 1,
}

var followLogsAndCommit = action.Action{
	Name: "follow-logs-and-commit",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(runContainerActionsArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		c, ok := ctx.Previous.(container.Container)
		if !ok {
			return nil, errors.New("Previous result must be a container.")
		}
		type logsResult struct {
			status int
			err    error
		}
		doneCh := make(chan bool)
		canceledCh := make(chan error)
		resultCh := make(chan logsResult)
		go func() {
			for {
				err := checkCanceled(args.event)
				if err != nil {
					select {
					case <-doneCh:
					case canceledCh <- err:
					}
					return
				}
				select {
				case <-doneCh:
					return
				case <-time.After(time.Second):
				}
			}
		}()
		go func() {
			status, err := c.Logs(args.provisioner, args.writer)
			select {
			case resultCh <- logsResult{status: status, err: err}:
			default:
			}
		}()
		select {
		case err := <-canceledCh:
			return nil, err
		case result := <-resultCh:
			doneCh <- true
			if result.err != nil {
				log.Errorf("error on get logs for container %s - %s", c.ID, result.err)
				return nil, result.err
			}
			if result.status != 0 {
				return nil, errors.Errorf("Exit status %d", result.status)
			}
		}
		fmt.Fprintf(args.writer, "\n---- Building application image ----\n")
		imageId, err := c.Commit(args.provisioner, args.writer)
		if err != nil {
			log.Errorf("error on commit container %s - %s", c.ID, err)
			return nil, err
		}
		fmt.Fprintf(args.writer, " ---> Cleaning up\n")
		c.Remove(args.provisioner)
		return imageId, nil
	},
	Backward: func(ctx action.BWContext) {
	},
	MinParams: 1,
}

var updateAppImage = action.Action{
	Name: "update-app-image",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		args := ctx.Params[0].(changeUnitsPipelineArgs)
		if err := checkCanceled(args.event); err != nil {
			return nil, err
		}
		currentImageName, _ := image.AppCurrentImageName(args.app.GetName())
		if currentImageName != args.imageId {
			err := image.AppendAppImageName(args.app.GetName(), args.imageId)
			if err != nil {
				return nil, errors.Wrap(err, "unable to save image name")
			}
		}
		imgHistorySize := image.ImageHistorySize()
		allImages, err := image.ListAppImages(args.app.GetName())
		if err != nil {
			log.Errorf("Couldn't list images for cleaning: %s", err)
			return ctx.Previous, nil
		}
		for i, imgName := range allImages {
			if i > len(allImages)-imgHistorySize-1 {
				err := args.provisioner.Cluster().RemoveImageIgnoreLast(imgName)
				if err != nil {
					log.Debugf("Ignored error removing old image %q: %s", imgName, err.Error())
				}
				continue
			}
			args.provisioner.cleanImage(args.app.GetName(), imgName)
		}
		return ctx.Previous, nil
	},
	OnError: rollbackNotice,
}
