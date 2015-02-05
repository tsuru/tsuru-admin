// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/provision"
	"github.com/tsuru/tsuru/provision/provisiontest"
	"launchpad.net/gocheck"
)

func (s *S) TestLogRemoveIsRegistered(c *gocheck.C) {
	manager := buildManager("tsuru-admin")
	token, ok := manager.Commands["log-remove"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(token, gocheck.FitsTypeOf, &logRemove{})
}

func (s *S) TestPlatformAddIsRegistered(c *gocheck.C) {
	manager := buildManager("tsuru-admin")
	token, ok := manager.Commands["platform-add"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(token, gocheck.FitsTypeOf, &platformAdd{})
}

func (s *S) TestCommandsFromBaseManagerAreRegistered(c *gocheck.C) {
	baseManager := cmd.BuildBaseManager("tsuru", version, header, nil)
	manager := buildManager("tsuru")
	for name, instance := range baseManager.Commands {
		command, ok := manager.Commands[name]
		c.Assert(ok, gocheck.Equals, true)
		c.Assert(command, gocheck.FitsTypeOf, instance)
	}
}

func (s *S) TestShouldRegisterAllCommandsFromProvisioners(c *gocheck.C) {
	fp := provisiontest.NewFakeProvisioner()
	p := AdminCommandableProvisioner{FakeProvisioner: *fp}
	provision.Register("fakeAdminProvisioner", &p)
	manager := buildManager("tsuru-admin")
	fake, ok := manager.Commands["fake-admin"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(fake, gocheck.FitsTypeOf, &FakeAdminCommand{})
}

func (s *S) TestViewUserQuotaIsRegistered(c *gocheck.C) {
	manager := buildManager("tsuru-admin")
	viewQuota, ok := manager.Commands["view-user-quota"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(viewQuota, gocheck.FitsTypeOf, viewUserQuota{})
}

func (s *S) TestChangeUserQuotaIsRegistered(c *gocheck.C) {
	manager := buildManager("tsuru-admin")
	changeQuota, ok := manager.Commands["change-user-quota"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(changeQuota, gocheck.FitsTypeOf, changeUserQuota{})
}

func (s *S) TestViewAppQuotaIsRegistered(c *gocheck.C) {
	manager := buildManager("tsuru-admin")
	viewQuota, ok := manager.Commands["view-app-quota"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(viewQuota, gocheck.FitsTypeOf, viewAppQuota{})
}

func (s *S) TestChangeAppQuotaIsRegistered(c *gocheck.C) {
	manager := buildManager("tsuru-admin")
	changeQuota, ok := manager.Commands["change-app-quota"]
	c.Assert(ok, gocheck.Equals, true)
	c.Assert(changeQuota, gocheck.FitsTypeOf, changeAppQuota{})
}
