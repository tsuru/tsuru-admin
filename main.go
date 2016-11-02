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
	m.Register(&admin.MachineList{})
	m.Register(&admin.MachineDestroy{})
	m.RegisterRemoved("app-unlock", "You should use `tsuru app-unlock` instead.")
	m.RegisterDeprecated(&userQuotaView{}, "view-user-quota")
	m.RegisterDeprecated(&userChangeQuota{}, "change-user-quota")
	m.RegisterDeprecated(&appQuotaView{}, "view-app-quota")
	m.RegisterDeprecated(&appQuotaChange{}, "change-app-quota")
	m.RegisterRemoved("plan-create", "You should use `tsuru plan-create` instead.")
	m.RegisterRemoved("plan-remove", "You should use `tsuru plan-remove` instead.")
	m.RegisterRemoved("router-list", "You should use `tsuru router-list` instead.")
	m.Register(&admin.TemplateList{})
	m.Register(&admin.TemplateAdd{})
	m.Register(&admin.TemplateRemove{})
	m.RegisterRemoved("user-list", "You should use `tsuru user-list` instead.")
	m.RegisterDeprecated(&admin.AddPoolToSchedulerCmd{}, "docker-pool-add")
	m.Register(&updatePoolToSchedulerCmd{})
	m.RegisterDeprecated(&removePoolFromSchedulerCmd{}, "docker-pool-remove")
	m.RegisterRemoved("pool-list", "You should use `tsuru pool-list` instead.")
	m.RegisterDeprecated(addTeamsToPoolCmd{}, "docker-pool-teams-add")
	m.RegisterDeprecated(removeTeamsFromPoolCmd{}, "docker-pool-teams-remove")
	m.Register(&cmd.ShellToContainerCmd{})
	m.RegisterRemoved("app-routes-rebuild", "You should use `tsuru app-routes-rebuild` instead.")
	m.Register(&admin.TemplateUpdate{})
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
