// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/auth"
	"github.com/tsuru/tsuru/cmd"
	"launchpad.net/gnuflag"
)

type tokenGen struct {
	export bool
}

func (c *tokenGen) Run(ctx *cmd.Context, client *cmd.Client) error {
	app := ctx.Args[0]
	url, err := cmd.GetURL("/tokens")
	if err != nil {
		return err
	}
	body := strings.NewReader(fmt.Sprintf(`{"client":"%s","export":%v}`, app, c.export))
	request, _ := http.NewRequest("POST", url, body)
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var token map[string]string
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return err
	}
	fmt.Fprintf(ctx.Stdout, "Application token: %q.\n", token["token"])
	return nil
}

func (c *tokenGen) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "token-gen",
		MinArgs: 1,
		Usage:   "token-gen <app-name>",
		Desc:    "Generates an authentication token for an app.",
	}
}

func (c *tokenGen) Flags() *gnuflag.FlagSet {
	fs := gnuflag.NewFlagSet("token-gen", gnuflag.ExitOnError)
	fs.BoolVar(&c.export, "export", false, "Define the token as environment variable in the app")
	fs.BoolVar(&c.export, "e", false, "Define the token as environment variable in the app")
	return fs
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
	var users []auth.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return err
	}
	println(users)
	table := cmd.NewTable()
	table.Headers = cmd.Row([]string{"User", "Teams"})
	for _, u := range users {
		teams, err := u.Teams()
		if err != nil {
			return err
		}
		teams_name := auth.GetTeamsNames(teams)
		table.AddRow(cmd.Row([]string{u.Email, strings.Join(teams_name, ", ")}))
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
