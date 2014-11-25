// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/iaas"
)

type machineList struct{}

func (c *machineList) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machine-list",
		Usage:   "machine-list",
		Desc:    "List all machines created using a IaaS.",
		MinArgs: 0,
	}
}

func (c *machineList) Run(context *cmd.Context, client *cmd.Client) error {
	url, err := cmd.GetURL("/iaas/machines")
	if err != nil {
		return err
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	var machines []iaas.Machine
	err = json.NewDecoder(response.Body).Decode(&machines)
	if err != nil {
		return err
	}
	table := cmd.NewTable()
	table.Headers = cmd.Row([]string{"Id", "IaaS", "Address", "Creation Params"})
	table.LineSeparator = true
	for _, machine := range machines {
		var params []string
		for k, v := range machine.CreationParams {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(params)
		table.AddRow(cmd.Row([]string{machine.Id, machine.Iaas, machine.Address, strings.Join(params, "\n")}))
	}
	table.Sort()
	context.Stdout.Write(table.Bytes())
	return nil
}

type machineDestroy struct{}

func (c *machineDestroy) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machine-destroy",
		Usage:   "machine-destroy <machine id>",
		Desc:    "Destroy an existing machine created using a IaaS.",
		MinArgs: 1,
	}
}
func (c *machineDestroy) Run(context *cmd.Context, client *cmd.Client) error {
	machineId := context.Args[0]
	url, err := cmd.GetURL("/iaas/machines/" + machineId)
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
	fmt.Fprintln(context.Stdout, "Machine successfully destroyed.")
	return nil
}

type templateList struct{}

func (c *templateList) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machine-template-list",
		Usage:   "machine-template-list",
		Desc:    "List all machine templates.",
		MinArgs: 0,
	}
}

func (c *templateList) Run(context *cmd.Context, client *cmd.Client) error {
	url, err := cmd.GetURL("/iaas/templates")
	if err != nil {
		return err
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	var templates []iaas.Template
	err = json.NewDecoder(response.Body).Decode(&templates)
	if err != nil {
		return err
	}
	table := cmd.NewTable()
	table.Headers = cmd.Row([]string{"Name", "IaaS", "Params"})
	table.LineSeparator = true
	for _, template := range templates {
		var params []string
		for _, data := range template.Data {
			params = append(params, fmt.Sprintf("%s=%s", data.Name, data.Value))
		}
		sort.Strings(params)
		table.AddRow(cmd.Row([]string{template.Name, template.IaaSName, strings.Join(params, "\n")}))
	}
	table.Sort()
	context.Stdout.Write(table.Bytes())
	return nil
}
