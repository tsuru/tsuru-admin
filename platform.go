// Copyright 2015 tsuru-admin authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/tsuru/gnuflag"
	"github.com/tsuru/tsuru/cmd"
)

type platformAdd struct {
	name       string
	dockerfile string
	image      string
	fs         *gnuflag.FlagSet
}

func (p *platformAdd) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "platform-add",
		Usage: "platform-add <platform name> [--dockerfile/-d Dockerfile] [--image/-i image]",
		Desc: `Adds new platform to tsuru.

The name of the image can be automatically inferred in case you're using an official platform. Check https://github.com/tsuru/platforms for a list of official platforms.

Examples:

	tsuru-admin platform-add java # uses the default Java image
	tsuru-admin platform-add java -i registry.company.com/tsuru/java # uses custom Java image
	tsuru-admin platform-add java -d /data/projects/java/Dockerfile # uses local Dockerfile
	tsuru-admin platform-add java -d https://platforms.com/java/Dockerfile #uses remote Dockerfile`,
		MinArgs: 1,
	}
}

func (p *platformAdd) Run(context *cmd.Context, client *cmd.Client) error {
	context.RawOutput()
	var body bytes.Buffer
	writer, err := serializeDockerfile(context.Args[0], &body, p.dockerfile, p.image, true)
	if err != nil {
		return err
	}
	writer.WriteField("name", context.Args[0])
	writer.Close()
	url, err := cmd.GetURL("/platforms")
	request, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return cmd.StreamJSONResponse(context.Stdout, response)
}

func (p *platformAdd) Flags() *gnuflag.FlagSet {
	dockerfileMessage := "URL or path to the Dockerfile used for building the image of the platform"
	if p.fs == nil {
		p.fs = gnuflag.NewFlagSet("platform-add", gnuflag.ExitOnError)
		p.fs.StringVar(&p.dockerfile, "dockerfile", "", dockerfileMessage)
		p.fs.StringVar(&p.dockerfile, "d", "", dockerfileMessage)
		p.fs.StringVar(&p.image, "image", "", "Name of the prebuilt Docker image")
		p.fs.StringVar(&p.image, "i", "", "Name of the prebuilt Docker image")
	}
	return p.fs
}

type platformUpdate struct {
	name        string
	dockerfile  string
	image       string
	forceUpdate bool
	disable     bool
	enable      bool
	fs          *gnuflag.FlagSet
}

func (p *platformUpdate) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "platform-update",
		Usage: "platform-update <platform name> [--dockerfile/-d Dockerfile] [--disable/--enable] [--image/-i image]",
		Desc: `Update a platform in tsuru."

The name of the image can be automatically inferred in case you're using an official platform. Check https://github.com/tsuru/platforms for a list of official platforms.

The flags --enable and --disable can be used for enabling or disabling a platform.

Examples:

	tsuru-admin platform-update java # uses the default Java image
	tsuru-admin platform-update java -i registry.company.com/tsuru/java # uses custom Java image
	tsuru-admin platform-update java -d /data/projects/java/Dockerfile # uses local Dockerfile
	tsuru-admin platform-update java -d https://platforms.com/java/Dockerfile #uses remote Dockerfile`,
		MinArgs: 1,
	}
}

func (p *platformUpdate) Flags() *gnuflag.FlagSet {
	dockerfileMessage := "URL or path to the Dockerfile used for building the image of the platform"
	if p.fs == nil {
		p.fs = gnuflag.NewFlagSet("platform-update", gnuflag.ExitOnError)
		p.fs.StringVar(&p.dockerfile, "dockerfile", "", dockerfileMessage)
		p.fs.StringVar(&p.dockerfile, "d", "", dockerfileMessage)
		p.fs.BoolVar(&p.disable, "disable", false, "Disable the platform")
		p.fs.BoolVar(&p.enable, "enable", false, "Enable the platform")
		p.fs.StringVar(&p.image, "image", "", "Name of the prebuilt Docker image")
		p.fs.StringVar(&p.image, "i", "", "Name of the prebuilt Docker image")
	}
	return p.fs
}

func (p *platformUpdate) Run(context *cmd.Context, client *cmd.Client) error {
	context.RawOutput()
	name := context.Args[0]
	if p.disable && p.enable {
		return errors.New("Conflicting options: --enable and --disable")
	}
	var disable string
	if p.enable {
		disable = "false"
	}
	if p.disable {
		disable = "true"
	}
	var body bytes.Buffer
	implicitImage := !p.disable && !p.enable && p.dockerfile == "" && p.image == ""
	writer, err := serializeDockerfile(context.Args[0], &body, p.dockerfile, p.image, implicitImage)
	if err != nil {
		return err
	}
	writer.WriteField("disabled", disable)
	writer.Close()
	url, err := cmd.GetURL(fmt.Sprintf("/platforms/%s", name))
	request, err := http.NewRequest("PUT", url, &body)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return cmd.StreamJSONResponse(context.Stdout, response)
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

func serializeDockerfile(name string, w io.Writer, dockerfile, image string, useImplicit bool) (*multipart.Writer, error) {
	if dockerfile != "" && image != "" {
		return nil, errors.New("Conflicting options: --image and --dockerfile")
	}
	writer := multipart.NewWriter(w)
	var dockerfileContent []byte
	if image != "" {
		dockerfileContent = []byte("FROM " + image)
	} else if dockerfile != "" {
		dockerfileURL, err := url.Parse(dockerfile)
		if err != nil {
			return nil, err
		}
		switch dockerfileURL.Scheme {
		case "http", "https":
			dockerfileContent, err = downloadDockerfile(dockerfile)
		default:
			dockerfileContent, err = ioutil.ReadFile(dockerfile)
		}
		if err != nil {
			return nil, err
		}
	} else if useImplicit {
		dockerfileContent = []byte("FROM tsuru/" + name)
	} else {
		return writer, nil
	}
	fileWriter, err := writer.CreateFormFile("dockerfile_content", "Dockerfile")
	if err != nil {
		return nil, err
	}
	fileWriter.Write(dockerfileContent)
	return writer, nil
}

func downloadDockerfile(dockerfileURL string) ([]byte, error) {
	resp, err := http.Get(dockerfileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
