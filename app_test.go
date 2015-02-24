// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"gopkg.in/check.v1"
)

func (s *S) TestAppLockDeleteInfo(c *check.C) {
	expected := &cmd.Info{
		Name:  "app-unlock",
		Usage: "app-unlock -a <app-name> [-y]",
		Desc: `Forces the removal of an app lock.
Use with caution, removing an active lock may cause inconsistencies.`,
		MinArgs: 0,
	}
	c.Assert((&appLockDelete{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestAppLockDeleteRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	expected := "Lock successfully removed!\n"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/apps/app1/lock" && req.Method == "DELETE"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := appLockDelete{}
	command.Flags().Parse(true, []string{"--app", "app1", "-y"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestAppLockDeleteRunAsksConfirmation(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Stdin:  strings.NewReader("n\n"),
	}
	command := appLockDelete{}
	command.Flags().Parse(true, []string{"--app", "app1"})
	err := command.Run(&context, nil)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Are you sure you want to remove the lock from app \"app1\"? (y/n) Abort.\n")
}
