// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/testing"
	"io/ioutil"
	"launchpad.net/gocheck"
	"net/http"
)

func (s *S) TestViewUserQuotaInfo(c *gocheck.C) {
	expected := &cmd.Info{
		Name:    "view-user-quota",
		MinArgs: 1,
		Usage:   "view-user-quota <user-email>",
		Desc:    "Displays the current usage and limit of the user",
	}
	c.Assert(viewUserQuota{}.Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestViewUserQuotaRun(c *gocheck.C) {
	result := `{"inuse":3,"limit":4}`
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: result, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.Method == "GET" && req.URL.Path == "/users/fss@corp.globo.com/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	expected := `User: fss@corp.globo.com
Apps usage: 3/4
`
	c.Assert(stdout.String(), gocheck.Equals, expected)
}

func (s *S) TestViewUserQuotaRunFailure(c *gocheck.C) {
	context := cmd.Context{Args: []string{"fss@corp.globo.com"}}
	trans := testing.Transport{Message: "user not found", Status: http.StatusNotFound}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "user not found")
}

func (s *S) TestChangeUserQuotaInfo(c *gocheck.C) {
	desc := `Changes the limit of apps that a user can create

The new limit must be an integer, it may also be "unlimited".`
	expected := &cmd.Info{
		Name:    "change-user-quota",
		MinArgs: 2,
		Usage:   "change-user-quota <user-email> <new-limit>",
		Desc:    desc,
	}
	c.Assert(changeUserQuota{}.Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestChangeUserQuotaRun(c *gocheck.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com", "5"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, gocheck.IsNil)
			c.Assert(string(body), gocheck.Equals, `limit=5`)
			return req.Method == "POST" && req.URL.Path == "/users/fss@corp.globo.com/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, "Quota successfully updated.\n")
	c.Assert(called, gocheck.Equals, true)
}

func (s *S) TestChangeUserQuotaRunUnlimited(c *gocheck.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com", "unlimited"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, gocheck.IsNil)
			c.Assert(string(body), gocheck.Equals, "limit=-1")
			c.Assert(req.Header.Get("Content-Type"), gocheck.Equals, "application/x-www-form-urlencoded")
			return req.Method == "POST" && req.URL.Path == "/users/fss@corp.globo.com/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, "Quota successfully updated.\n")
	c.Assert(called, gocheck.Equals, true)
}

func (s *S) TestChangeUserQuotaRunInvalidLimit(c *gocheck.C) {
	context := cmd.Context{Args: []string{"fss@corp.globo.com", "unlimiteddd"}}
	command := changeUserQuota{}
	err := command.Run(&context, nil)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `invalid limit. It must be either an integer or "unlimited"`)
}

func (s *S) TestChangeUserQuotaFailure(c *gocheck.C) {
	var stdout, stderr bytes.Buffer
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := &testing.Transport{
		Message: "user not found",
		Status:  http.StatusNotFound,
	}
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com", "5"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := changeUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "user not found")
}

func (s *S) TestViewAppQuotaInfo(c *gocheck.C) {
	expected := &cmd.Info{
		Name:    "view-app-quota",
		MinArgs: 1,
		Usage:   "view-app-quota <app-name>",
		Desc:    "Displays the current usage and limit of the given app",
	}
	c.Assert(viewAppQuota{}.Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestViewAppQuotaRun(c *gocheck.C) {
	result := `{"inuse":3,"limit":4}`
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"hibria"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: result, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.Method == "GET" && req.URL.Path == "/apps/hibria/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	expected := `App: hibria
Units usage: 3/4
`
	c.Assert(stdout.String(), gocheck.Equals, expected)
}

func (s *S) TestViewAppQuotaRunFailure(c *gocheck.C) {
	context := cmd.Context{Args: []string{"hybria"}}
	trans := testing.Transport{Message: "app not found", Status: http.StatusNotFound}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "app not found")
}

func (s *S) TestChangeAppQuotaInfo(c *gocheck.C) {
	desc := `Changes the limit of units that an app can have

The new limit must be an integer, it may also be "unlimited".`
	expected := &cmd.Info{
		Name:    "change-app-quota",
		MinArgs: 2,
		Usage:   "change-app-quota <user-email> <new-limit>",
		Desc:    desc,
	}
	c.Assert(changeAppQuota{}.Info(), gocheck.DeepEquals, expected)
}

func (s *S) TestChangeAppQuotaRun(c *gocheck.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"myapp", "5"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, gocheck.IsNil)
			c.Assert(string(body), gocheck.Equals, `limit=5`)
			return req.Method == "POST" && req.URL.Path == "/apps/myapp/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, "Quota successfully updated.\n")
	c.Assert(called, gocheck.Equals, true)
}

func (s *S) TestChangeAppQuotaRunUnlimited(c *gocheck.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"myapp", "unlimited"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := testing.ConditionalTransport{
		Transport: testing.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, gocheck.IsNil)
			c.Assert(string(body), gocheck.Equals, "limit=-1")
			c.Assert(req.Header.Get("Content-Type"), gocheck.Equals, "application/x-www-form-urlencoded")
			return req.Method == "POST" && req.URL.Path == "/apps/myapp/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.IsNil)
	c.Assert(stdout.String(), gocheck.Equals, "Quota successfully updated.\n")
	c.Assert(called, gocheck.Equals, true)
}

func (s *S) TestChangeAppQuotaRunInvalidLimit(c *gocheck.C) {
	context := cmd.Context{Args: []string{"myapp", "unlimiteddd"}}
	command := changeAppQuota{}
	err := command.Run(&context, nil)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, `invalid limit. It must be either an integer or "unlimited"`)
}

func (s *S) TestChangeAppQuotaFailure(c *gocheck.C) {
	var stdout, stderr bytes.Buffer
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := &testing.Transport{
		Message: "app not found",
		Status:  http.StatusNotFound,
	}
	context := cmd.Context{
		Args:   []string{"myapp", "5"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, manager)
	command := changeAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "app not found")
}
