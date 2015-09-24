// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"github.com/tsuru/tsuru/io"
	"gopkg.in/check.v1"
)

func (s *S) TestPlatformAddInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "platform-add",
		Usage:   "platform-add <platform name> [--dockerfile/-d Dockerfile]",
		Desc:    "Add new platform to tsuru.",
		MinArgs: 1,
	}

	c.Assert((&platformAdd{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestPlatformAddRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"teste"},
	}
	expected := "\nOK!\nPlatform successfully added!\n"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "\nOK!\n", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			return req.URL.Path == "/platforms" && req.Method == "POST"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformAdd{}
	command.Flags().Parse(true, []string{"--dockerfile", "http://localhost/Dockerfile"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestPlatformAddFlagSet(c *check.C) {
	message := "The dockerfile url to create a platform"
	command := platformAdd{}
	flagset := command.Flags()
	flagset.Parse(true, []string{"--dockerfile", "dockerfile"})

	dockerfile := flagset.Lookup("dockerfile")
	c.Check(dockerfile.Name, check.Equals, "dockerfile")
	c.Check(dockerfile.Usage, check.Equals, message)
	c.Check(dockerfile.DefValue, check.Equals, "")

	sdockerfile := flagset.Lookup("d")
	c.Check(sdockerfile.Name, check.Equals, "d")
	c.Check(sdockerfile.Usage, check.Equals, message)
	c.Check(sdockerfile.DefValue, check.Equals, "")
}

func (s *S) TestPlatformUpdateInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "platform-update",
		Usage:   "platform-update <platform name> [--dockerfile/-d Dockerfile] [--disable/--enable]",
		Desc:    "Update a platform to tsuru.",
		MinArgs: 1,
	}
	c.Assert((&platformUpdate{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestPlatformUpdateFlagSet(c *check.C) {
	dockerfile_message := "The dockerfile url to update a platform"
	command := platformUpdate{}
	flagset := command.Flags()
	flagset.Parse(true, []string{"--dockerfile", "dockerfile"})

	dockerfile := flagset.Lookup("dockerfile")
	c.Check(dockerfile.Name, check.Equals, "dockerfile")
	c.Check(dockerfile.Usage, check.Equals, dockerfile_message)
	c.Check(dockerfile.DefValue, check.Equals, "")

	sdockerfile := flagset.Lookup("d")
	c.Check(sdockerfile.Name, check.Equals, "d")
	c.Check(sdockerfile.Usage, check.Equals, dockerfile_message)
	c.Check(sdockerfile.DefValue, check.Equals, "")
}

func (s *S) TestPlatformUpdateRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	name := "teste"
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{name},
	}
	expectedMsg := "--something--\nPlatform successfully updated!\n"
	msg := io.SimpleJsonMessage{Message: expectedMsg}
	result, err := json.Marshal(msg)
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: string(result), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			c.Assert(req.FormValue("dockerfile"), check.Equals, "http://localhost/Dockerfile")
			c.Assert(req.URL.Path, check.Equals, "/platforms/"+name)
			c.Assert(req.Method, check.Equals, "PUT")
			return req.URL.Path == "/platforms/"+name && req.Method == "PUT"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformUpdate{}
	command.Flags().Parse(true, []string{"--dockerfile", "http://localhost/Dockerfile"})
	err = command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expectedMsg)
}

func (s *S) TestPlatformUpdateWithFlagDisableTrue(c *check.C) {
	var stdout, stderr bytes.Buffer
	name := "teste"
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{name},
	}
	expectedMsg := "Platform successfully updated!\n"
	msg := io.SimpleJsonMessage{Message: expectedMsg}
	result, err := json.Marshal(msg)
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: string(result), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			c.Assert(req.URL.Path, check.Equals, "/platforms/"+name)
			c.Assert(req.Method, check.Equals, "PUT")
			return req.URL.Path == "/platforms/"+name && req.Method == "PUT" && req.URL.RawQuery == "disabled=true"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformUpdate{}
	command.Flags().Parse(true, []string{"--disable"})
	err = command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expectedMsg)
}

func (s *S) TestPlatformUpdateWithFlagEnabledTrue(c *check.C) {
	var stdout, stderr bytes.Buffer
	name := "teste"
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{name},
	}
	expectedMsg := "Platform successfully updated!\n"
	msg := io.SimpleJsonMessage{Message: expectedMsg}
	result, err := json.Marshal(msg)
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: string(result), Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			c.Assert(req.URL.Path, check.Equals, "/platforms/"+name)
			c.Assert(req.Method, check.Equals, "PUT")
			return req.URL.Path == "/platforms/"+name && req.Method == "PUT" && req.URL.RawQuery == "disabled=false"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformUpdate{}
	command.Flags().Parse(true, []string{"--enable"})
	err = command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expectedMsg)
}

func (s *S) TestPlatformUpdateWithWrongFlag(c *check.C) {
	var stdout, stderr bytes.Buffer
	name := "teste"
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{name},
	}
	expected := "Conflicting options: --enable and --disable\n"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			return req.URL.Path == "/platforms/"+name && req.Method == "PUT" && req.URL.RawQuery == "disabled=true"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformUpdate{}
	command.Flags().Parse(true, []string{"--disable", "--enable"})
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, expected)
}

func (s *S) TestPlatformUpdateWithError(c *check.C) {
	var stdout, stderr bytes.Buffer
	name := "teste"
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{name},
	}
	expectedError := "Flag is required"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			return req.URL.Path == "/platforms/"+name && req.Method == "PUT"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformUpdate{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, expectedError)
}

func (s *S) TestPlatformRemoveRun(c *check.C) {
	var stdout, stderr bytes.Buffer
	name := "teste"
	context := cmd.Context{
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{name},
	}
	expected := "Platform successfully removed!\n"
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/platforms/"+name && req.Method == "DELETE"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := platformRemove{}
	command.Flags().Parse(true, []string{"-y"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestPlatformRemoveInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "platform-remove",
		Usage:   "platform-remove <platform name> [-y]",
		Desc:    "Remove a platform from tsuru.",
		MinArgs: 1,
	}
	c.Assert((&platformRemove{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestPlatformRemoveConfirmation(c *check.C) {
	var stdout bytes.Buffer
	context := cmd.Context{
		Stdout: &stdout,
		Stdin:  strings.NewReader("n\n"),
		Args:   []string{"something"},
	}
	command := platformRemove{}
	err := command.Run(&context, nil)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Are you sure you want to remove \"something\" platform? (y/n) Abort.\n")
}
