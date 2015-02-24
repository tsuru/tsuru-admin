// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"gopkg.in/check.v1"
)

func (s *S) TestLogRemoveInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "log-remove",
		Usage:   "log-remove [--app appname]",
		Desc:    `remove all app logs.`,
		MinArgs: 0,
	}
	c.Assert((&logRemove{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestLogRemoveRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	expected := "Logs successfully removed!\n"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/logs" && req.Method == "DELETE"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := logRemove{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestLogRemoveByAppRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	expected := "Logs successfully removed!\n"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/logs" && req.Method == "DELETE" && req.URL.RawQuery == "app=app1"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := logRemove{}
	command.Flags().Parse(true, []string{"--app", "app1"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestLogRemoveFlagSet(c *check.C) {
	command := logRemove{}
	flagset := command.Flags()
	flagset.Parse(true, []string{"--app", "ashamed"})
	app := flagset.Lookup("app")
	c.Check(app, check.NotNil)
	c.Check(app.Name, check.Equals, "app")
	c.Check(app.Usage, check.Equals, "The name of the app.")
	c.Check(app.Value.String(), check.Equals, "ashamed")
	c.Check(app.DefValue, check.Equals, "")
	sapp := flagset.Lookup("a")
	c.Check(sapp, check.NotNil)
	c.Check(sapp.Name, check.Equals, "a")
	c.Check(sapp.Usage, check.Equals, "The name of the app.")
	c.Check(sapp.Value.String(), check.Equals, "ashamed")
	c.Check(sapp.DefValue, check.Equals, "")
}
