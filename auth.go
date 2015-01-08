// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/cmd"
)

type user struct {
	Email string
	Teams []string
}

type listUsers struct{}

func (c *listUsers) Run(ctx *cmd.Context, client *cmd.Client) error {
	url, err := cmd.GetURL("/users")
	if err != nil {
		return err
	}
	request, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var users []user
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return err
	}
	table := cmd.NewTable()
	table.Headers = cmd.Row([]string{"User", "Teams"})
	for _, u := range users {
		table.AddRow(cmd.Row([]string{u.Email, strings.Join(u.Teams, ", ")}))
	}
	table.LineSeparator = true
	table.Sort()
	ctx.Stdout.Write(table.Bytes())
	return nil
}

func (c *listUsers) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "user-list",
		MinArgs: 0,
		Usage:   "user-list",
		Desc:    "List all users in tsuru.",
	}
}
