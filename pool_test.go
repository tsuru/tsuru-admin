// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"github.com/tsuru/tsuru/provision"
	"gopkg.in/check.v1"
)

func (s *S) TestAddPoolToSchedulerCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:  "pool-add",
		Usage: "pool-add <pool> [-p/--public] [-d/--default] [-f/--force]",
		Desc: `Add a pool to cluster.
Use [-p/--public] flag to create a public pool.
Use [-d/--default] flag to create default pool.
Use [-f/--force] flag to force overwrite default pool.`,
		MinArgs: 1,
	}
	cmd := addPoolToSchedulerCmd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestAddPoolToTheSchedulerCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"poolTest"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := addPoolToSchedulerCmd{}
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestAddPublicPool(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"test"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"name":    "test",
				"public":  true,
				"default": false,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := addPoolToSchedulerCmd{}
	cmd.Flags().Parse(true, []string{"-p"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestAddDefaultPool(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"test"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"name":    "test",
				"public":  false,
				"default": true,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := addPoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"-d"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestFailToAddMoreThanOneDefaultPool(c *check.C) {
	var buf bytes.Buffer
	stdin := bytes.NewBufferString("no")
	transportError := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusPreconditionFailed, Message: "Default pool already exist."},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"name":    "test",
				"public":  false,
				"default": true,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool"
		},
	}
	manager := cmd.Manager{}
	context := cmd.Context{Args: []string{"test"}, Stdout: &buf, Stdin: stdin}
	client := cmd.NewClient(&http.Client{Transport: &transportError}, nil, &manager)
	command := addPoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"-d"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := "WARNING: Default pool already exist. Do you want change to test pool? (y/n) Pool add aborted.\n"
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestForceToOverwriteDefaultPool(c *check.C) {
	var buf bytes.Buffer
	stdin := bytes.NewBufferString("no")
	transportError := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusPreconditionFailed, Message: "Default pool already exist."},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"name":    "test",
				"public":  false,
				"default": true,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(result, check.DeepEquals, expected)
			return req.URL.RawQuery == "force=true"
		},
	}
	manager := cmd.Manager{}
	context := cmd.Context{Args: []string{"test"}, Stdout: &buf, Stdin: stdin}
	client := cmd.NewClient(&http.Client{Transport: &transportError}, nil, &manager)
	command := addPoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"-d"})
	command.Flags().Parse(true, []string{"-f"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestAskOverwriteDefaultPool(c *check.C) {
	var buf bytes.Buffer
	var called int
	stdin := bytes.NewBufferString("yes")
	transportError := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusPreconditionFailed, Message: "Default pool already exist."},
		CondFunc: func(req *http.Request) bool {
			called += 1
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"name":    "test",
				"public":  false,
				"default": true,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.RawQuery == "force=false"
		},
	}
	transportOk := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusOK, Message: ""},
		CondFunc: func(req *http.Request) bool {
			called += 1
			return req.URL.RawQuery == "force=true"
		},
	}
	multiTransport := cmdtest.MultiConditionalTransport{
		ConditionalTransports: []cmdtest.ConditionalTransport{transportError, transportOk},
	}
	context := cmd.Context{
		Args:   []string{"test"},
		Stdout: &buf,
		Stdin:  stdin,
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: &multiTransport}, nil, &manager)
	command := addPoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"-d"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(called, check.Equals, 2)
	expected := "WARNING: Default pool already exist. Do you want change to test pool? (y/n) Pool successfully registered.\n"
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestUpdatePoolToSchedulerCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:  "pool-update",
		Usage: "pool-update <pool> [--public=true/false] [--default=true/false] [-f/--force]",
		Desc: `Update a pool.
Use [--public=true/false] to change the pool attribute.
Use [--default=true/false] to change the pool attribute.
Use [-f/--force] to force pool to be default.`,
		MinArgs: 1,
	}
	cmd := updatePoolToSchedulerCmd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestUpdatePoolToTheSchedulerCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"poolTest"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool/poolTest"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := updatePoolToSchedulerCmd{}
	cmd.Flags().Parse(true, []string{"--public", "true"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestFailToUpdateMoreThanOneDefaultPool(c *check.C) {
	var buf bytes.Buffer
	stdin := bytes.NewBufferString("no")
	transportError := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusPreconditionFailed, Message: "Default pool already exist."},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"Public":  false,
				"Default": true,
				"Force":   false,
				"Name":    "",
				"NewName": "",
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool/test"
		},
	}
	manager := cmd.Manager{}
	context := cmd.Context{Args: []string{"test"}, Stdout: &buf, Stdin: stdin}
	client := cmd.NewClient(&http.Client{Transport: &transportError}, nil, &manager)
	command := updatePoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"--default=true"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := "WARNING: Default pool already exist. Do you want change to test pool? (y/n) Pool update aborted.\n"
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestForceToOverwriteDefaultPoolInUpdate(c *check.C) {
	var buf bytes.Buffer
	stdin := bytes.NewBufferString("no")
	transportError := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusPreconditionFailed, Message: "Default pool already exist."},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"Public":  false,
				"Default": true,
				"Force":   true,
				"Name":    "",
				"NewName": "",
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(result, check.DeepEquals, expected)
			return req.URL.Path == "/pool/test"
		},
	}
	manager := cmd.Manager{}
	context := cmd.Context{Args: []string{"test"}, Stdout: &buf, Stdin: stdin}
	client := cmd.NewClient(&http.Client{Transport: &transportError}, nil, &manager)
	command := updatePoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"--default=true"})
	command.Flags().Parse(true, []string{"-f"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestAskOverwriteDefaultPoolInUpdate(c *check.C) {
	var buf bytes.Buffer
	var called int
	stdin := bytes.NewBufferString("yes")
	transportError := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusPreconditionFailed, Message: "Default pool already exist."},
		CondFunc: func(req *http.Request) bool {
			called += 1
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"Public":  false,
				"Default": true,
				"Force":   false,
				"Name":    "",
				"NewName": "",
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool/test"
		},
	}
	transportOk := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusOK, Message: ""},
		CondFunc: func(req *http.Request) bool {
			called += 1
			return req.URL.Path == "/pool/test"
		},
	}
	multiTransport := cmdtest.MultiConditionalTransport{
		ConditionalTransports: []cmdtest.ConditionalTransport{transportError, transportOk},
	}
	context := cmd.Context{
		Args:   []string{"test"},
		Stdout: &buf,
		Stdin:  stdin,
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: &multiTransport}, nil, &manager)
	command := updatePoolToSchedulerCmd{}
	command.Flags().Parse(true, []string{"--default=true"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(called, check.Equals, 2)
	expected := "WARNING: Default pool already exist. Do you want change to test pool? (y/n) Pool successfully updated.\n"
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestRemovePoolFromSchedulerCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:    "pool-remove",
		Usage:   "pool-remove <pool> [-y]",
		Desc:    "Remove a pool to cluster",
		MinArgs: 1,
	}
	cmd := removePoolFromSchedulerCmd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestRemovePoolFromTheSchedulerCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"poolTest"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	cmd := removePoolFromSchedulerCmd{}
	cmd.Flags().Parse(true, []string{"-y"})
	err := cmd.Run(&context, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestRemovePoolFromTheSchedulerCmdConfirmation(c *check.C) {
	var stdout bytes.Buffer
	context := cmd.Context{
		Args:   []string{"poolX"},
		Stdout: &stdout,
		Stdin:  strings.NewReader("n\n"),
	}
	command := removePoolFromSchedulerCmd{}
	err := command.Run(&context, nil)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Are you sure you want to remove \"poolX\" pool? (y/n) Abort.\n")
}

func (s *S) TestListPoolsInTheSchedulerCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:  "pool-list",
		Usage: "pool-list",
		Desc:  "List available pools in the cluster",
	}
	cmd := listPoolsInTheSchedulerCmd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestListPoolsInTheSchedulerCmdRun(c *check.C) {
	var buf bytes.Buffer
	pool := provision.Pool{Name: "pool1", Teams: []string{"tsuruteam", "ateam"}}
	pools := []provision.Pool{pool}
	poolsJson, _ := json.Marshal(pools)
	ctx := cmd.Context{Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: string(poolsJson), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := listPoolsInTheSchedulerCmd{}.Run(&ctx, client)
	c.Assert(err, check.IsNil)
	expected := `+-------+------------------+
| Pools | Teams            |
+-------+------------------+
| pool1 | tsuruteam, ateam |
+-------+------------------+
`
	c.Assert(buf.String(), check.Equals, expected)
}

func (s *S) TestAddTeamsToPoolCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:    "pool-teams-add",
		Usage:   "pool-teams-add <pool> <teams>",
		Desc:    "Add team to a pool",
		MinArgs: 2,
	}
	cmd := addTeamsToPoolCmd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestAddTeamsToPoolCmdRun(c *check.C) {
	var buf bytes.Buffer
	ctx := cmd.Context{Stdout: &buf, Args: []string{"pool1", "team1", "team2"}}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool/team"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := addTeamsToPoolCmd{}.Run(&ctx, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestRemoveTeamsFromPoolCmdInfo(c *check.C) {
	expected := cmd.Info{
		Name:    "pool-teams-remove",
		Usage:   "pool-teams-remove <pool> <teams>",
		Desc:    "Remove team from pool",
		MinArgs: 2,
	}
	cmd := removeTeamsFromPoolCmd{}
	c.Assert(cmd.Info(), check.DeepEquals, &expected)
}

func (s *S) TestRemoveTeamsFromPoolCmdRun(c *check.C) {
	var buf bytes.Buffer
	ctx := cmd.Context{Stdout: &buf, Args: []string{"pool1", "team1"}}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool/team"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := removeTeamsFromPoolCmd{}.Run(&ctx, client)
	c.Assert(err, check.IsNil)
}
