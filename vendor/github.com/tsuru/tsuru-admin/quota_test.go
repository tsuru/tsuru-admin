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

func (s *S) TestUserQuotaViewInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "user-quota-view",
		MinArgs: 1,
		Usage:   "user-quota-view <user-email>",
		Desc:    "Displays the current usage and limit of the user",
	}
	c.Assert((&userQuotaView{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestUserQuotaViewRun(c *check.C) {
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
	command := userQuotaView{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `User: fss@corp.globo.com
Apps usage: 3/4
`
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestUserQuotaViewRunFailure(c *check.C) {
	context := cmd.Context{Args: []string{"fss@corp.globo.com"}}
	trans := cmdtest.Transport{Message: "user not found", Status: http.StatusNotFound}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := userQuotaView{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "user not found")
}

func (s *S) TestUserChangeQuotaInfo(c *check.C) {
	desc := `Changes the limit of apps that a user can create

The new limit must be an integer, it may also be "unlimited".`
	expected := &cmd.Info{
		Name:    "user-quota-change",
		MinArgs: 2,
		Usage:   "user-quota-change <user-email> <new-limit>",
		Desc:    desc,
	}
	c.Assert((&userChangeQuota{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestUserChangeQuotaRun(c *check.C) {
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
	command := userChangeQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestUserChangeQuotaRunUnlimited(c *check.C) {
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
	command := userChangeQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestUserChangeQuotaRunInvalidLimit(c *check.C) {
	context := cmd.Context{Args: []string{"fss@corp.globo.com", "unlimiteddd"}}
	command := userChangeQuota{}
	err := command.Run(&context, nil)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, `invalid limit. It must be either an integer or "unlimited"`)
}

func (s *S) TestUserChangeQuotaFailure(c *check.C) {
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
	command := userChangeQuota{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "user not found")
}

func (s *S) TestAppQuotaViewInfo(c *check.C) {
	expected := &cmd.Info{
		Name:    "app-quota-view",
		MinArgs: 1,
		Usage:   "app-quota-view <app-name>",
		Desc:    "Displays the current usage and limit of the given app",
	}
	c.Assert((&appQuotaView{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestAppQuotaViewRun(c *check.C) {
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
	command := appQuotaView{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	expected := `App: hibria
Units usage: 3/4
`
	c.Assert(stdout.String(), check.Equals, expected)
}

func (s *S) TestAppQuotaViewRunFailure(c *check.C) {
	context := cmd.Context{Args: []string{"hybria"}}
	trans := cmdtest.Transport{Message: "app not found", Status: http.StatusNotFound}
	client := cmd.NewClient(&http.Client{Transport: &trans}, nil, manager)
	command := appQuotaView{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "app not found")
}

func (s *S) TestAppQuotaChangeInfo(c *check.C) {
	desc := `Changes the limit of units that an app can have

The new limit must be an integer, it may also be "unlimited".`
	expected := &cmd.Info{
		Name:    "app-quota-change",
		MinArgs: 2,
		Usage:   "app-quota-change <app-name> <new-limit>",
		Desc:    desc,
	}
	c.Assert((&appQuotaChange{}).Info(), check.DeepEquals, expected)
}

func (s *S) TestAppQuotaChangeRun(c *check.C) {
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
	command := appQuotaChange{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestAppQuotaChangeRunUnlimited(c *check.C) {
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
	command := appQuotaChange{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(stdout.String(), check.Equals, "Quota successfully updated.\n")
	c.Assert(called, check.Equals, true)
}

func (s *S) TestAppQuotaChangeRunInvalidLimit(c *check.C) {
	context := cmd.Context{Args: []string{"myapp", "unlimiteddd"}}
	command := appQuotaChange{}
	err := command.Run(&context, nil)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, `invalid limit. It must be either an integer or "unlimited"`)
}

func (s *S) TestAppQuotaChangeFailure(c *check.C) {
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
	command := appQuotaChange{}
	err := command.Run(&context, client)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "app not found")
}
