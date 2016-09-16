// Copyright 2016 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/tsuru/tsuru-client/tsuru/admin"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/provision"
	_ "github.com/tsuru/tsuru/provision/docker"
)

const (
	version = "1.0.0"
	header  = "Supported-Tsuru-Admin"
)

func buildManager(name string) *cmd.Manager {
	m := cmd.BuildBaseManager(name, version, header, nil)
	m.RegisterRemoved("log-remove", "This action is no longer supported.")
	m.Register(&admin.PlatformAdd{})
	m.Register(&platformUpdate{})
	m.Register(&platformRemove{})
	m.Register(&machineList{})
	m.Register(&machineDestroy{})
	m.Register(&appLockDelete{})
	m.RegisterDeprecated(&userQuotaView{}, "view-user-quota")
	m.RegisterDeprecated(&userChangeQuota{}, "change-user-quota")
	m.RegisterDeprecated(&appQuotaView{}, "view-app-quota")
	m.RegisterDeprecated(&appQuotaChange{}, "change-app-quota")
	m.Register(&planCreate{})
	m.Register(&planRemove{})
	m.Register(&planRoutersList{})
	m.Register(&templateList{})
	m.Register(&templateAdd{})
	m.Register(&templateRemove{})
	m.RegisterRemoved("user-list", "You should use `tsuru user-list` instead.")
	m.RegisterDeprecated(&admin.AddPoolToSchedulerCmd{}, "docker-pool-add")
	m.Register(&updatePoolToSchedulerCmd{})
	m.RegisterDeprecated(&removePoolFromSchedulerCmd{}, "docker-pool-remove")
	m.RegisterRemoved("pool-list", "You should use `tsuru pool-list` instead.")
	m.RegisterDeprecated(addTeamsToPoolCmd{}, "docker-pool-teams-add")
	m.RegisterDeprecated(removeTeamsFromPoolCmd{}, "docker-pool-teams-remove")
	m.Register(&cmd.ShellToContainerCmd{})
	m.Register(&appRoutesRebuild{})
	m.Register(&templateUpdate{})
	m.RegisterDeprecated(&admin.AddNodeCmd{}, "docker-node-add")
	m.RegisterDeprecated(&admin.RemoveNodeCmd{}, "docker-node-remove")
	m.RegisterDeprecated(&admin.UpdateNodeCmd{}, "docker-node-update")
	m.RegisterDeprecated(&admin.ListNodesCmd{}, "docker-node-list")
	registerProvisionersCommands(m)
	return m
}

func registerProvisionersCommands(m *cmd.Manager) {
	provisioners, err := provision.Registry()
	if err != nil {
		log.Fatalf("unable to load provisioners: %s", err)
	}
	for _, p := range provisioners {
		if c, ok := p.(cmd.AdminCommandable); ok {
			commands := c.AdminCommands()
			for _, cmd := range commands {
				m.Register(cmd)
			}
		}
	}
}

func main() {
	name := cmd.ExtractProgramName(os.Args[0])
	manager := buildManager(name)
	args := os.Args[1:]
	manager.Run(args)
}
