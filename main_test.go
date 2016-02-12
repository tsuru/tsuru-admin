// Copyright 2016 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/provisiontest"
	"gopkg.in/check.v1"
)

func (s *S) TestPlatformAddIsRegistered(c *check.C) {
	manager := buildManager("tsuru-admin")
	token, ok := manager.Commands["platform-add"]
	c.Assert(ok, check.Equals, true)
	c.Assert(token, check.FitsTypeOf, &platformAdd{})
}

func (s *S) TestCommandsFromBaseManagerAreRegistered(c *check.C) {
	baseManager := cmd.BuildBaseManager("tsuru", version, header, nil)
	manager := buildManager("tsuru")
	for name, instance := range baseManager.Commands {
		command, ok := manager.Commands[name]
		c.Assert(ok, check.Equals, true)
		c.Assert(command, check.FitsTypeOf, instance)
	}
}

func (s *S) TestShouldRegisterAllCommandsFromProvisioners(c *check.C) {
	fp := provisiontest.NewFakeProvisioner()
	p := AdminCommandableProvisioner{FakeProvisioner: *fp}
	provision.Register("fakeAdminProvisioner", &p)
	manager := buildManager("tsuru-admin")
	fake, ok := manager.Commands["fake-admin"]
	c.Assert(ok, check.Equals, true)
	c.Assert(fake, check.FitsTypeOf, &FakeAdminCommand{})
}

func (s *S) TestUserQuotaViewIsRegistered(c *check.C) {
	manager := buildManager("tsuru-admin")
	viewQuota, ok := manager.Commands["user-quota-view"]
	c.Assert(ok, check.Equals, true)
	c.Assert(viewQuota, check.FitsTypeOf, &userQuotaView{})
}

func (s *S) TestUserChangeQuotaIsRegistered(c *check.C) {
	manager := buildManager("tsuru-admin")
	changeQuota, ok := manager.Commands["user-quota-change"]
	c.Assert(ok, check.Equals, true)
	c.Assert(changeQuota, check.FitsTypeOf, &userChangeQuota{})
}

func (s *S) TestAppQuotaViewIsRegistered(c *check.C) {
	manager := buildManager("tsuru-admin")
	viewQuota, ok := manager.Commands["app-quota-view"]
	c.Assert(ok, check.Equals, true)
	c.Assert(viewQuota, check.FitsTypeOf, &appQuotaView{})
}

func (s *S) TestAppQuotaChangeIsRegistered(c *check.C) {
	manager := buildManager("tsuru-admin")
	changeQuota, ok := manager.Commands["app-quota-change"]
	c.Assert(ok, check.Equals, true)
	c.Assert(changeQuota, check.FitsTypeOf, &appQuotaChange{})
}
