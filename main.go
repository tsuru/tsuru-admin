// Copyright 2014 tsuru-admin authors. All rights reserved.
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
	version = "0.8.2"
	header  = "Supported-Tsuru-Admin"
)

func buildManager(name string) *cmd.Manager {
	m := cmd.BuildBaseManager(name, version, header, nil)
	m.Register(&tokenGen{})
	m.Register(&logRemove{})
	m.Register(&platformAdd{})
	m.Register(&platformUpdate{})
	m.Register(&platformRemove{})
	m.RegisterDeprecated(&machineList{}, "machines-list")
	m.Register(&machineDestroy{})
	m.Register(&appLockDelete{})
	m.Register(viewUserQuota{})
	m.Register(changeUserQuota{})
	m.Register(viewAppQuota{})
	m.Register(changeAppQuota{})
	m.Register(&planCreate{})
	m.Register(&planRemove{})
	m.Register(&templateList{})
	m.Register(&templateAdd{})
	m.Register(&templateRemove{})
	m.Register(&listUsers{})
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
