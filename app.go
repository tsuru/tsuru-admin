// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/tsuru-base"
	"launchpad.net/gnuflag"
)

type appLockDelete struct {
	tsuru.GuessingCommand
	tsuru.ConfirmationCommand
}

func (c *appLockDelete) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "app-unlock",
		MinArgs: 0,
		Usage:   "app-unlock -a <app-name> [-y]",
		Desc: `Forces the removal of an app lock.
Use with caution, removing an active lock may cause inconsistencies.`,
	}
}

func (c *appLockDelete) Run(ctx *cmd.Context, client *cmd.Client) error {
	appName, err := c.Guess()
	if err != nil {
		return err
	}
	if !c.Confirm(ctx, fmt.Sprintf(`Are you sure you want to remove the lock from app "%s"?`, appName)) {
		return nil
	}
	url, err := cmd.GetURL("/apps/" + appName + "/lock")
	if err != nil {
		return err
	}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		return err
	}
	fmt.Fprintf(ctx.Stdout, "Lock successfully removed!\n")
	return nil
}

func (c *appLockDelete) Flags() *gnuflag.FlagSet {
	return cmd.MergeFlagSet(
		c.GuessingCommand.Flags(),
		c.ConfirmationCommand.Flags(),
	)
}
