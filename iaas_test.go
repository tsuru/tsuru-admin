// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"github.com/tsuru/tsuru/iaas"
	"gopkg.in/check.v1"
)

func (s *S) TestMachineListInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "machine-list",
		Usage:   "machine-list",
		Desc:    "List all machines created using a IaaS.",
		MinArgs: 0,
	}
	c.Assert((&machineList{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestMachineListRun(c *check.C) {
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
	c.Assert(err, check.IsNil)
	expected := `+-----+-------+---------+-----------------+
| Id  | IaaS  | Address | Creation Params |
+-----+-------+---------+-----------------+
| id1 | iaas1 | addr1   | param1=value1   |
+-----+-------+---------+-----------------+
| id2 | iaas2 | addr2   | param1=value1   |
|     |       |         | param2=value2   |
+-----+-------+---------+-----------------+
`
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: string(data), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/iaas/machines" && req.Method == "GET"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := machineList{}
	err = command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestMachineDestroyInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "machine-destroy",
		Usage:   "machine-destroy <machine id>",
		Desc:    "Destroy an existing machine created using a IaaS.",
		MinArgs: 1,
	}
	c.Assert((&machineDestroy{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestMachineDestroyRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"myid1"},
	}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/iaas/machines/myid1" && req.Method == "DELETE"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := machineDestroy{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Machine successfully destroyed.\n")
}

func (s *S) TestTemplateListInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "machine-template-list",
		Usage:   "machine-template-list",
		Desc:    "List all machine templates.",
		MinArgs: 0,
	}
	c.Assert((&templateList{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestTemplateListRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	tpl1 := iaas.Template{Name: "machine1", IaaSName: "ec2", Data: iaas.TemplateDataList{
		{Name: "region", Value: "us-east-1"},
		{Name: "type", Value: "m1.small"},
	}}
	tpl2 := iaas.Template{Name: "tpl1", IaaSName: "ec2", Data: iaas.TemplateDataList{
		{Name: "region", Value: "xxxx"},
		{Name: "type", Value: "l1.large"},
	}}
	data, err := json.Marshal([]iaas.Template{tpl1, tpl2})
	c.Assert(err, check.IsNil)
	expected := `+----------+------+------------------+
| Name     | IaaS | Params           |
+----------+------+------------------+
| machine1 | ec2  | region=us-east-1 |
|          |      | type=m1.small    |
+----------+------+------------------+
| tpl1     | ec2  | region=xxxx      |
|          |      | type=l1.large    |
+----------+------+------------------+
`
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: string(data), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/iaas/templates" && req.Method == "GET"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := templateList{}
	err = command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestTemplateAddCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:    "machine-template-add",
		Usage:   "machine-template-add <name> <iaas> <param>=<value>...",
		Desc:    "Add a new machine template.",
		MinArgs: 3,
	}
	cmd := templateAdd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestTemplateAddCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"my-tpl", "ec2", "zone=xyz", "image=ami-something"}, Stdout: &buf}
	expectedBody := `{"Name":"my-tpl","IaaSName":"ec2",` +
		`"Data":[{"Name":"zone","Value":"xyz"},{"Name":"image","Value":"ami-something"}]}`
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			body, _ := ioutil.ReadAll(req.Body)
			c.Assert(string(body), check.DeepEquals, expectedBody)
			return req.URL.Path == "/iaas/templates" && req.Method == "POST"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := templateAdd{}
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Template successfully added.\n")
}

func (s *S) TestTemplateRemoveCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:    "machine-template-remove",
		Usage:   "machine-template-remove <name>",
		Desc:    "Remove an existing machine template.",
		MinArgs: 1,
	}
	cmd := templateRemove{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestTemplateRemoveCmdRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"my-tpl"}, Stdout: &buf, Stderr: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/iaas/templates/my-tpl" && req.Method == "DELETE"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := templateRemove{}
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Template successfully removed.\n")
}
