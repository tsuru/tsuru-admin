// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"gopkg.in/check.v1"
)

func (s *S) TestViewUserQuotaInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "view-user-quota",
		MinArgs: 1,
		Usage:   "view-user-quota <user-email>",
		Desc:    "Displays the current usage and limit of the user",
	}
	c.Assert(viewUserQuota{}.Info(), check.DeepEquals, expected)
}

func (s *S) TestViewUserQuotaRun(c *check.C) {
	result := `{"inuse":3,"limit":4}`
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: result, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.Method == "GET" && req.URL.Path == "/users/fss@corp.globo.com/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `User: fss@corp.globo.com
Apps usage: 3/4
`
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestViewUserQuotaRunFailure(c *check.C) {
	context := cmd.Context{Args: []string{"fss@corp.globo.com"}}
	trans := cmdtest.Transport{Message: "user not found", Status: http.StatusNotFound}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "user not found")
}

func (s *S) TestChangeUserQuotaInfo(c *check.C) {
	desc := `Changes the limit of apps that a user can create

The new limit must be an integer, it may also be "unlimited".`
	expected := &cmd.Info{
		Name:    "change-user-quota",
		MinArgs: 2,
		Usage:   "change-user-quota <user-email> <new-limit>",
		Desc:    desc,
	}
	c.Assert(changeUserQuota{}.Info(), check.DeepEquals, expected)
}

func (s *S) TestChangeUserQuotaRun(c *check.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com", "5"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			c.Assert(string(body), check.Equals, `limit=5`)
			return req.Method == "POST" && req.URL.Path == "/users/fss@corp.globo.com/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestChangeUserQuotaRunUnlimited(c *check.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"fss@corp.globo.com", "unlimited"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			c.Assert(string(body), check.Equals, "limit=-1")
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			return req.Method == "POST" && req.URL.Path == "/users/fss@corp.globo.com/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeUserQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestChangeUserQuotaRunInvalidLimit(c *check.C) {
	context := cmd.Context{Args: []string{"fss@corp.globo.com", "unlimiteddd"}}
	command := changeUserQuota{}
	err := command.Run(&context, nil)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, `invalid limit. It must be either an integer or "unlimited"`)
}

func (s *S) TestChangeUserQuotaFailure(c *check.C) {
	var stdout, stderr bytes.Buffer
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := &cmdtest.Transport{
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
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "user not found")
}

func (s *S) TestViewAppQuotaInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "view-app-quota",
		MinArgs: 1,
		Usage:   "view-app-quota <app-name>",
		Desc:    "Displays the current usage and limit of the given app",
	}
	c.Assert(viewAppQuota{}.Info(), check.DeepEquals, expected)
}

func (s *S) TestViewAppQuotaRun(c *check.C) {
	result := `{"inuse":3,"limit":4}`
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"hibria"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: result, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.Method == "GET" && req.URL.Path == "/apps/hibria/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `App: hibria
Units usage: 3/4
`
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestViewAppQuotaRunFailure(c *check.C) {
	context := cmd.Context{Args: []string{"hybria"}}
	trans := cmdtest.Transport{Message: "app not found", Status: http.StatusNotFound}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := viewAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "app not found")
}

func (s *S) TestChangeAppQuotaInfo(c *check.C) {
	desc := `Changes the limit of units that an app can have

The new limit must be an integer, it may also be "unlimited".`
	expected := &cmd.Info{
		Name:    "change-app-quota",
		MinArgs: 2,
		Usage:   "change-app-quota <app-name> <new-limit>",
		Desc:    desc,
	}
	c.Assert(changeAppQuota{}.Info(), check.DeepEquals, expected)
}

func (s *S) TestChangeAppQuotaRun(c *check.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"myapp", "5"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			c.Assert(string(body), check.Equals, `limit=5`)
			return req.Method == "POST" && req.URL.Path == "/apps/myapp/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestChangeAppQuotaRunUnlimited(c *check.C) {
	var called bool
	var stdout, stderr bytes.Buffer
	context := cmd.Context{
		Args:   []string{"myapp", "unlimited"},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			called = true
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			c.Assert(err, check.IsNil)
			c.Assert(string(body), check.Equals, "limit=-1")
			c.Assert(req.Header.Get("Content-Type"), check.Equals, "application/x-www-form-urlencoded")
			return req.Method == "POST" && req.URL.Path == "/apps/myapp/quota"
		},
	}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := changeAppQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestChangeAppQuotaRunInvalidLimit(c *check.C) {
	context := cmd.Context{Args: []string{"myapp", "unlimiteddd"}}
	command := changeAppQuota{}
	err := command.Run(&context, nil)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, `invalid limit. It must be either an integer or "unlimited"`)
}

func (s *S) TestChangeAppQuotaFailure(c *check.C) {
	var stdout, stderr bytes.Buffer
	manager := cmd.NewManager("tsuru", "0.5", "ad-ver", &stdout, &stderr, nil, nil)
	trans := &cmdtest.Transport{
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
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "app not found")
}
