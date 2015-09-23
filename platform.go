// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tsuru/tsuru/cmd"
	"launchpad.net/gnuflag"
)

type platformAdd struct {
	name       string
	dockerfile string
	fs         *gnuflag.FlagSet
}

func (p *platformAdd) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "platform-add",
		Usage:   "platform-add <platform name> [--dockerfile/-d Dockerfile]",
		Desc:    "Add new platform to tsuru.",
		MinArgs: 1,
	}
}

func (p *platformAdd) Run(context *cmd.Context, client *cmd.Client) error {
	context.RawOutput()
	name := context.Args[0]
	body := fmt.Sprintf("name=%s&dockerfile=%s", name, p.dockerfile)
	url, err := cmd.GetURL("/platforms")
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var buf bytes.Buffer
	for n := int64(1); n > 0 && err == nil; n, err = io.Copy(io.MultiWriter(&buf, context.Stdout), response.Body) {
	}
	if strings.HasSuffix(buf.String(), "\nOK!\n") {
		fmt.Fprintf(context.Stdout, "Platform successfully added!\n")
		return nil
	}
	return errors.New("Failed to add new platform.\n")
}

func (p *platformAdd) Flags() *gnuflag.FlagSet {
	message := "The dockerfile url to create a platform"
	if p.fs == nil {
		p.fs = gnuflag.NewFlagSet("platform-add", gnuflag.ExitOnError)
		p.fs.StringVar(&p.dockerfile, "dockerfile", "", message)
		p.fs.StringVar(&p.dockerfile, "d", "", message)
	}
	return p.fs
}

type platformUpdate struct {
	name        string
	dockerfile  string
	forceUpdate bool
	disabled    bool
	enabled     bool
	fs          *gnuflag.FlagSet
}

func (p *platformUpdate) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "platform-update",
		Usage:   "platform-update <platform name> [--dockerfile/-d Dockerfile] [--disable/--enable]",
		Desc:    "Update a platform to tsuru.",
		MinArgs: 1,
	}
}

func (p *platformUpdate) Flags() *gnuflag.FlagSet {
	dockerfileMessage := "The dockerfile url to update a platform"
	if p.fs == nil {
		p.fs = gnuflag.NewFlagSet("platform-update", gnuflag.ExitOnError)
		p.fs.StringVar(&p.dockerfile, "dockerfile", "", dockerfileMessage)
		p.fs.StringVar(&p.dockerfile, "d", "", dockerfileMessage)
		p.fs.BoolVar(&p.disabled, "disable", false, "Disable the platform")
		p.fs.BoolVar(&p.enabled, "enable", false, "Enable the platform")
	}
	return p.fs
}

//colocar dockerfile certo no metodo acima
func (p *platformUpdate) Run(context *cmd.Context, client *cmd.Client) error {
	context.RawOutput()
	name := context.Args[0]
	if p.disabled && p.enabled {
		return errors.New("Conflicting options: --enable and --disable\n")
	}
	disableVar := ""
	if p.enabled {
		disableVar = "false"
	}
	if p.disabled {
		disableVar = "true"
	}
	body := fmt.Sprintf("a=1&dockerfile=%s", p.dockerfile)
	url, err := cmd.GetURL(fmt.Sprintf("/platforms/%s?disabled=%s", name, disableVar))
	request, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var buf bytes.Buffer
	for n := int64(1); n > 0 && err == nil; n, err = io.Copy(io.MultiWriter(&buf, context.Stdout), response.Body) {
	}
	if strings.HasSuffix(buf.String(), "\nOK!\n") {
		fmt.Fprintf(context.Stdout, "Platform successfully updated!\n")
		return nil
	}
	return errors.New("Failed to update platform!\n")
}

type platformRemove struct {
	cmd.ConfirmationCommand
}

func (p *platformRemove) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "platform-remove",
		Usage:   "platform-remove <platform name> [-y]",
		Desc:    "Remove a platform from tsuru.",
		MinArgs: 1,
	}
}

func (p *platformRemove) Run(context *cmd.Context, client *cmd.Client) error {
	name := context.Args[0]
	if !p.Confirm(context, fmt.Sprintf(`Are you sure you want to remove "%s" platform?`, name)) {
		return nil
	}
	url, err := cmd.GetURL("/platforms/" + name)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		fmt.Fprintf(context.Stdout, "Failed to remove platform!\n")
		return err
	}
	fmt.Fprintf(context.Stdout, "Platform successfully removed!\n")
	return nil
}
