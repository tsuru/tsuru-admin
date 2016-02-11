// Copyright 2014 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
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

type templateAdd struct{}

func (c *templateAdd) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machine-template-add",
		Usage:   "machine-template-add <name> <iaas> <param>=<value>...",
		Desc:    "Add a new machine template.",
		MinArgs: 3,
	}
}

func (c *templateAdd) Run(context *cmd.Context, client *cmd.Client) error {
	var template iaas.Template
	template.Name = context.Args[0]
	template.IaaSName = context.Args[1]
	for _, param := range context.Args[2:] {
		if strings.Contains(param, "=") {
			keyValue := strings.SplitN(param, "=", 2)
			template.Data = append(template.Data, iaas.TemplateData{
				Name:  keyValue[0],
				Value: keyValue[1],
			})
		}
	}
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return err
	}
	url, err := cmd.GetURL("/iaas/templates")
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(templateBytes))
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		context.Stderr.Write([]byte("Failed to add template.\n"))
		return err
	}
	context.Stdout.Write([]byte("Template successfully added.\n"))
	return nil
}

type templateRemove struct{}

func (c *templateRemove) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machine-template-remove",
		Usage:   "machine-template-remove <name>",
		Desc:    "Remove an existing machine template.",
		MinArgs: 1,
	}
}

func (c *templateRemove) Run(context *cmd.Context, client *cmd.Client) error {
	url, err := cmd.GetURL("/iaas/templates/" + context.Args[0])
	if err != nil {
		return err
	}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		context.Stderr.Write([]byte("Failed to remove template.\n"))
		return err
	}
	context.Stdout.Write([]byte("Template successfully removed.\n"))
	return nil
}

type templateUpdate struct{}

func (c *templateUpdate) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "machine-template-update",
		Usage:   "machine-template-update <name> <param>=<value>...",
		Desc:    "Update an existing machine template.",
		MinArgs: 2,
	}
}

func (c *templateUpdate) Run(context *cmd.Context, client *cmd.Client) error {
	var template iaas.Template
	template.Name = context.Args[0]
	for _, param := range context.Args[1:] {
		if strings.Contains(param, "=") {
			keyValue := strings.SplitN(param, "=", 2)
			template.Data = append(template.Data, iaas.TemplateData{
				Name:  keyValue[0],
				Value: keyValue[1],
			})
		}
	}
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return err
	}
	url, err := cmd.GetURL(fmt.Sprintf("/iaas/templates/%s", template.Name))
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(templateBytes))
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		context.Stderr.Write([]byte("Failed to update template.\n"))
		return err
	}
	context.Stdout.Write([]byte("Template successfully updated.\n"))
	return nil
}
