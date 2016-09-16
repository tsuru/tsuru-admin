// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/provision/provisiontest"
	"gopkg.in/check.v1"
)

type S struct {
	recover []string
	manager *cmd.Manager
}

func (s *S) SetUpSuite(c *check.C) {
	var stdout, stderr bytes.Buffer
	s.manager = cmd.NewManager("glb", version, header, &stdout, &stderr, os.Stdin, nil)
	os.Setenv("TSURU_TARGET", "http://localhost")
}

func (s *S) TearDownSuite(c *check.C) {
	os.Unsetenv("TSURU_TARGET")
}

var _ = check.Suite(&S{})

func Test(t *testing.T) { check.TestingT(t) }

type AdminCommandableProvisioner struct {
	*provisiontest.FakeProvisioner
}

func (p *AdminCommandableProvisioner) AdminCommands() []cmd.Command {
	return []cmd.Command{&FakeAdminCommand{}}
}

type FakeAdminCommand struct{}

func (c *FakeAdminCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "fake-admin",
		Usage: "fake usage",
		Desc:  "fake desc",
	}
}

func (c *FakeAdminCommand) Run(*cmd.Context, *cmd.Client) error {
	return nil
}
