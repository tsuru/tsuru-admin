// Copyright 2016 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

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
	registerMigrated := func(cmd string, newCmd string) {
		if newCmd == "" {
			newCmd = cmd
		}
		m.RegisterRemoved(cmd, fmt.Sprintf("You should use `tsuru %s` instead.", newCmd))
	}
	m.RegisterRemoved("log-remove", "This action is no longer supported.")
	registerProvisionersCommands(m)
	registerMigrated("app-shell", "")
	registerMigrated("platform-update", "")
	registerMigrated("platform-remove", "")
	registerMigrated("machine-list", "")
	registerMigrated("machine-destroy", "")
	registerMigrated("app-unlock", "")
	registerMigrated("plan-create", "")
	registerMigrated("plan-remove", "")
	registerMigrated("router-list", "")
	registerMigrated("user-list", "")
	registerMigrated("pool-list", "")
	registerMigrated("app-routes-rebuild", "")
	registerMigrated("platform-add", "")
	registerMigrated("machine-template-list", "")
	registerMigrated("machine-template-add", "")
	registerMigrated("machine-template-remove", "")
	registerMigrated("machine-template-update", "")
	registerMigrated("user-quota-view", "")
	registerMigrated("user-quota-change", "")
	registerMigrated("app-quota-view", "")
	registerMigrated("app-quota-change", "")
	registerMigrated("view-user-quota", "user-quota-view")
	registerMigrated("change-user-quota", "user-quota-change")
	registerMigrated("view-app-quota", "app-quota-view")
	registerMigrated("change-app-quota", "app-quota-change")
	registerMigrated("docker-node-add", "node-add")
	registerMigrated("docker-node-remove", "node-remove")
	registerMigrated("docker-node-update", "node-update")
	registerMigrated("docker-node-list", "node-list")
	registerMigrated("docker-pool-add", "pool-add")
	registerMigrated("pool-add", "")
	registerMigrated("pool-update", "")
	registerMigrated("docker-pool-remove", "pool-remove")
	registerMigrated("pool-remove", "")
	registerMigrated("docker-pool-teams-add", "pool-teams-add")
	registerMigrated("pool-teams-add", "")
	registerMigrated("docker-pool-teams-remove", "pool-teams-remove")
	registerMigrated("pool-teams-remove", "")
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
				name := cmd.Info().Name
				m.RegisterRemoved(name, fmt.Sprintf("You should use `tsuru %s` instead.", name))
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
