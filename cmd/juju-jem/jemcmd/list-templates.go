// Copyright 2015 Canonical Ltd.

package jemcmd

import (
	"fmt"

	"github.com/juju/cmd"
	"gopkg.in/errgo.v1"

	"github.com/CanonicalLtd/jem/params"
)

type listTemplatesCommand struct {
	commandBase
}

var listTemplatesDoc = `
The list-templates command lists available templates.
`

func (c *listTemplatesCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "list-templates",
		Purpose: "list templates",
		Doc:     listTemplatesDoc,
	}
}

func (c *listTemplatesCommand) Init(args []string) error {
	if len(args) != 0 {
		return errgo.Newf("got %d arguments, want none", len(args))
	}
	return nil
}

func (c *listTemplatesCommand) Run(ctxt *cmd.Context) error {
	client, err := c.newClient()
	if err != nil {
		return errgo.Mask(err)
	}
	defer client.Close()
	resp, err := client.ListTemplates(&params.ListTemplates{})
	if err != nil {
		return errgo.Mask(err)
	}
	for _, e := range resp.Templates {
		fmt.Fprintf(ctxt.Stdout, "%s\n", e.Path)
	}
	return nil
}