// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// tsuru-admin is under development.
package main

import (
	"os"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/provision"
	_ "github.com/tsuru/tsuru/provision/docker"
)

const (
	version = "0.12.0"
	header  = "Supported-Tsuru-Admin"
)

func buildManager(name string) *cmd.Manager {
	m := cmd.BuildBaseManager(name, version, header, nil)
	m.Register(&logRemove{})
	m.Register(&platformAdd{})
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
	m.RegisterDeprecated(&addPoolToSchedulerCmd{}, "docker-pool-add")
	m.Register(&updatePoolToSchedulerCmd{})
	m.RegisterDeprecated(&removePoolFromSchedulerCmd{}, "docker-pool-remove")
	m.RegisterDeprecated(listPoolsInTheSchedulerCmd{}, "docker-pool-list")
	m.RegisterDeprecated(addTeamsToPoolCmd{}, "docker-pool-teams-add")
	m.RegisterDeprecated(removeTeamsFromPoolCmd{}, "docker-pool-teams-remove")
	m.Register(&cmd.ShellToContainerCmd{})
	m.Register(&appRoutesRebuild{})
	m.Register(&templateUpdate{})
	registerProvisionersCommands(m)
	return m
}

func registerProvisionersCommands(m *cmd.Manager) {
	provisioners := provision.Registry()
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
