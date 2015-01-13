// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/testing"
	"launchpad.net/gocheck"
)

func (s *S) TestListUsersInfo(c *gocheck.C) {
	expected := &cmd.Info{
		Name:    "user-list",
		MinArgs: 0,
		Usage:   "user-list",
		Desc:    "List all users in tsuru.",
	}
	c.Assert((&listUsers{}).Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestListUsersRun(c *gocheck.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("glb", "0.2", "ad-ver", &stdout, &stderr, nil, nil)
	result := `[{"email": "test@test.com","teams":["team1", "team2", "team3"]}]`
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: result, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.Method == "GET" && req.URL.Path == "/users"
		},
	}
	expected := `+---------------+---------------------+
| User          | Teams               |
+---------------+---------------------+
| test@test.com | team1, team2, team3 |
+---------------+---------------------+
`
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := listUsers{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, expected)
}
