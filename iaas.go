// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/iaas"
	"net/http"
	"sort"
	"strings"
)

type machinesList struct{}

func (c *machinesList) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machines-list",
		Usage:   "machines-list",
		Desc:    "List all machines created using a IaaS.",
		MinArgs: 0,
	}
}

func (c *machinesList) Run(context *cmd.Context, client *cmd.Client) error {
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
