// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/errors"
)

type addPoolToSchedulerCmd struct {
	public       bool
	defaultPool  bool
	forceDefault bool
	fs           *gnuflag.FlagSet
}

func (addPoolToSchedulerCmd) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "pool-add",
		Usage: "pool-add <pool> [-p/--public] [-d/--default] [-f/--force]",
		Desc: `Add a pool to cluster.
Use [-p/--public] flag to create a public pool.
Use [-d/--default] flag to create default pool.
Use [-f/--force] flag to force overwrite default pool.`,
		MinArgs: 1,
	}
}

func (c *addPoolToSchedulerCmd) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("", gnuflag.ExitOnError)
		c.fs.BoolVar(&c.public, "public", false, "Make pool public")
		c.fs.BoolVar(&c.public, "p", false, "Make pool public")
		c.fs.BoolVar(&c.defaultPool, "default", false, "Make pool default")
		c.fs.BoolVar(&c.defaultPool, "d", false, "Make pool default")
		c.fs.BoolVar(&c.forceDefault, "force", false, "Force overwrite default pool.")
		c.fs.BoolVar(&c.forceDefault, "f", false, "Force overwrite default pool.")
	}
	return c.fs
}

func (c *addPoolToSchedulerCmd) Run(ctx *cmd.Context, client *cmd.Client) error {
	b, err := json.Marshal(map[string]interface{}{
		"name":    ctx.Args[0],
		"public":  c.public,
		"default": c.defaultPool,
	})
	if err != nil {
		return err
	}
	url, err := cmd.GetURL(fmt.Sprintf("/pool?force=%t", c.forceDefault))
	err = doRequest(client, url, b)
	if err != nil {
		if e, ok := err.(*errors.HTTP); ok && e.Code == http.StatusPreconditionFailed {
			retryMessage := "WARNING: Default pool already exist. Do you want change to %s pool? (y/n) "
			url, _ := cmd.GetURL(fmt.Sprintf("/pool?force=%t", true))
			successMessage := "Pool successfully registered.\n"
			failMessage := "Pool add aborted.\n"
			return confirmAction(ctx, client, url, b, retryMessage, failMessage, successMessage)
		}
		return err
	}
	ctx.Stdout.Write([]byte("Pool successfully registered.\n"))
	return nil
}

func doRequest(client *cmd.Client, url string, body []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func confirmAction(ctx *cmd.Context, client *cmd.Client, url string, body []byte, retryMessage, failMessage, successMessage string) error {
	var answer string
	fmt.Fprintf(ctx.Stdout, retryMessage, ctx.Args[0])
	fmt.Fscanf(ctx.Stdin, "%s", &answer)
	if answer == "y" || answer == "yes" {
		err := doRequest(client, url, body)
		if err != nil {
			return err
		}
		ctx.Stdout.Write([]byte(successMessage))
		return nil

	}
	ctx.Stdout.Write([]byte(failMessage))
	return nil
}

type pointerBoolFlag struct {
	value *bool
}

func (p *pointerBoolFlag) String() string {
	return fmt.Sprintf("%#v", p)
}

func (p *pointerBoolFlag) Set(value string) error {
	if value == "" {
		return nil
	}
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	p.value = &v
	return nil
}

type updatePoolToSchedulerCmd struct {
	public       pointerBoolFlag
	defaultPool  pointerBoolFlag
	forceDefault bool
	fs           *gnuflag.FlagSet
}

func (updatePoolToSchedulerCmd) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "pool-update",
		Usage: "pool-update <pool> [--public=true/false] [--default=true/false] [-f/--force]",
		Desc: `Update a pool.
Use [--public=true/false] to change the pool attribute.
Use [--default=true/false] to change the pool attribute.
Use [-f/--force] to force pool to be default.`,
		MinArgs: 1,
	}
}

func (c *updatePoolToSchedulerCmd) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("", gnuflag.ExitOnError)
		c.fs.Var(&c.public, "public", "Make pool public.")
		c.fs.Var(&c.defaultPool, "default", "Make pool default.")
		c.fs.BoolVar(&c.forceDefault, "force", false, "Force pool to be default.")
		c.fs.BoolVar(&c.forceDefault, "f", false, "Force pool to be default.")
	}
	return c.fs
}

func (c *updatePoolToSchedulerCmd) Run(ctx *cmd.Context, client *cmd.Client) error {
	opts := map[string]*bool{
		"public":  c.public.value,
		"default": c.defaultPool.value,
	}
	b, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	url, err := cmd.GetURL(fmt.Sprintf("/pool/%s?force=%t", ctx.Args[0], c.forceDefault))
	err = doRequest(client, url, b)
	if err != nil {
		if e, ok := err.(*errors.HTTP); ok && e.Code == http.StatusPreconditionFailed {
			retryMessage := "WARNING: Default pool already exist. Do you want change to %s pool? (y/n) "
			failMessage := "Pool update aborted.\n"
			successMessage := "Pool successfully updated.\n"
			url, err = cmd.GetURL(fmt.Sprintf("/pool/%s?force=%t", ctx.Args[0], true))
			return confirmAction(ctx, client, url, b, retryMessage, failMessage, successMessage)
		}
		return err
	}
	ctx.Stdout.Write([]byte("Pool successfully updated.\n"))
	return nil
}

type removePoolFromSchedulerCmd struct {
	cmd.ConfirmationCommand
}

func (c *removePoolFromSchedulerCmd) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "pool-remove",
		Usage:   "pool-remove <pool> [-y]",
		Desc:    "Remove a pool to cluster",
		MinArgs: 1,
	}
}

func (c *removePoolFromSchedulerCmd) Run(ctx *cmd.Context, client *cmd.Client) error {
	if !c.Confirm(ctx, fmt.Sprintf("Are you sure you want to remove \"%s\" pool?", ctx.Args[0])) {
		return nil
	}
	b, err := json.Marshal(map[string]string{"pool": ctx.Args[0]})
	if err != nil {
		return err
	}
	url, err := cmd.GetURL("/pool")
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	ctx.Stdout.Write([]byte("Pool successfully removed.\n"))
	return nil
}

type addTeamsToPoolCmd struct{}

func (addTeamsToPoolCmd) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "pool-teams-add",
		Usage:   "pool-teams-add <pool> <teams>",
		Desc:    "Add team to a pool",
		MinArgs: 2,
	}
}

func (addTeamsToPoolCmd) Run(ctx *cmd.Context, client *cmd.Client) error {
	body, err := json.Marshal(map[string]interface{}{"teams": ctx.Args[1:]})
	if err != nil {
		return err
	}
	url, err := cmd.GetURL(fmt.Sprintf("/pool/%s/team", ctx.Args[0]))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	ctx.Stdout.Write([]byte("Teams successfully registered.\n"))
	return nil
}

type removeTeamsFromPoolCmd struct{}

func (removeTeamsFromPoolCmd) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "pool-teams-remove",
		Usage:   "pool-teams-remove <pool> <teams>",
		Desc:    "Remove team from pool",
		MinArgs: 2,
	}
}

func (removeTeamsFromPoolCmd) Run(ctx *cmd.Context, client *cmd.Client) error {
	body, err := json.Marshal(map[string]interface{}{"pool": ctx.Args[0], "teams": ctx.Args[1:]})
	if err != nil {
		return err
	}
	url, err := cmd.GetURL(fmt.Sprintf("/pool/%s/team", ctx.Args[0]))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	ctx.Stdout.Write([]byte("Teams successfully removed.\n"))
	return nil
}
