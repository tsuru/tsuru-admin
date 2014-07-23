// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/testing"
	"github.com/tsuru/tsuru/iaas"
	"launchpad.net/gocheck"
	"net/http"
)

func (s *S) TestMachinesListInfo(c *gocheck.C) {
	expected := &cmd.Info{
		Name:    "machines-list",
		Usage:   "machines-list",
		Desc:    "List all machines created using a IaaS.",
		MinArgs: 0,
	}
	c.Assert((&machinesList{}).Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestMachinesListRun(c *gocheck.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	m1 := iaas.Machine{Id: "id1", Address: "addr1", Iaas: "iaas1", CreationParams: map[string]string{
		"param1": "value1",
	}}
	m2 := iaas.Machine{Id: "id2", Address: "addr2", Iaas: "iaas2", CreationParams: map[string]string{
		"param1": "value1",
		"param2": "value2",
	}}
	data, err := json.Marshal([]iaas.Machine{m1, m2})
	c.Assert(err, gocheck.IsNil)
	expected := `+-----+-------+---------+-----------------+
| Id  | IaaS  | Address | Creation Params |
+-----+-------+---------+-----------------+
| id1 | iaas1 | addr1   | param1=value1   |
+-----+-------+---------+-----------------+
| id2 | iaas2 | addr2   | param1=value1   |
|     |       |         | param2=value2   |
+-----+-------+---------+-----------------+
`
	trans := &testing.ConditionalTransport{
		Transport: testing.Transport{Message: string(data), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/iaas/machines" && req.Method == "GET"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := machinesList{}
	err = command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, expected)
}

func (s *S) TestMachineDestroyInfo(c *gocheck.C) {
	expected := &cmd.Info{
		Name:    "machine-destroy",
		Usage:   "machine-destroy <machine id>",
		Desc:    "Destroy an existing machine created using a IaaS.",
		MinArgs: 1,
	}
	c.Assert((&machineDestroy{}).Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestMachineDestroyRun(c *gocheck.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"myid1"},
	}
	trans := &testing.ConditionalTransport{
		Transport: testing.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/iaas/machines/myid1" && req.Method == "DELETE"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := machineDestroy{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, "Machine successfully destroyed.\n")
}
