// Copyright 2016 tsuru-admin authors. All rights reserved.
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
	"gopkg.in/check.v1"
)

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

func (s *S) TestUpdatePoolToTheSchedulerCmd(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"poolTest"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			expected := map[string]interface{}{
				"default": interface{}(nil),
				"public":  true,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool/poolTest" && req.URL.Query().Get("force") == "false"
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
				"default": true,
				"public":  interface{}(nil),
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool/test" && req.URL.Query().Get("force") == "false"
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
				"default": true,
				"public":  interface{}(nil),
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(result, check.DeepEquals, expected)
			return req.URL.Path == "/pool/test" && req.URL.Query().Get("force") == "true"
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
			a := new(bool)
			*a = true
			expected := map[string]interface{}{
				"default": true,
				"public":  nil,
			}
			result := map[string]interface{}{}
			err = json.Unmarshal(body, &result)
			c.Assert(expected, check.DeepEquals, result)
			return req.URL.Path == "/pool/test" && req.URL.Query().Get("force") == "false"
		},
	}
	transportOk := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Status: http.StatusOK, Message: ""},
		CondFunc: func(req *http.Request) bool {
			called += 1
			return req.URL.Path == "/pool/test" && req.URL.Query().Get("force") == "true"
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

func (s *S) TestAddTeamsToPoolCmdRun(c *check.C) {
	var buf bytes.Buffer
	ctx := cmd.Context{Stdout: &buf, Args: []string{"pool1", "team1", "team2"}}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool/pool1/team"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := addTeamsToPoolCmd{}.Run(&ctx, client)
	c.Assert(err, check.IsNil)
}

func (s *S) TestRemoveTeamsFromPoolCmdRun(c *check.C) {
	var buf bytes.Buffer
	ctx := cmd.Context{Stdout: &buf, Args: []string{"pool1", "team1"}}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/pool/pool1/team"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	err := removeTeamsFromPoolCmd{}.Run(&ctx, client)
	c.Assert(err, check.IsNil)
}
