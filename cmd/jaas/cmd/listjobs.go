// Copyright 2026 Canonical.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/canonical/jimm/v3/pkg/api/params"
	"github.com/juju/cmd/v3"
	"github.com/juju/gnuflag"
	jujucmd "github.com/juju/juju/cmd"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/juju/jujuclient"
)

const (
	listjobsCommandDoc = `
Displays controller information for all jobs known to JIMM.
`
	listjobsCommandExample = `
    juju jobs
    juju jobs --format json
`
)

// NewListjobsCommand returns a command to list controller information.
func NewListJobsCommand() cmd.Command {
	cmd := &listjobsCommand{}
	cmd.SetClientStore(jujuclient.NewFileClientStore())

	return modelcmd.WrapBase(cmd)
}

// listjobsCommand shows controller information
// for all jobs known to JIMM.
type listjobsCommand struct {
	jaasCommandBase
	out cmd.Output
}

func (c *listjobsCommand) Info() *cmd.Info {
	return jujucmd.Info(&cmd.Info{
		Name:     "jobs",
		Purpose:  "Lists all jobs known to JIMM.",
		Doc:      listjobsCommandDoc,
		Examples: listjobsCommandExample,
		Aliases:  []string{"list-jobs"},
	})
}

// SetFlags implements Command.SetFlags.
func (c *listjobsCommand) SetFlags(f *gnuflag.FlagSet) {
	c.CommandBase.SetFlags(f)
	c.out.AddFlags(f, "yaml", map[string]cmd.Formatter{
		"yaml": cmd.FormatYaml,
		"json": cmd.FormatJson,
	})
}

var jobs = []params.ListJobInfo{}
var i = 0

func init() {
	// This is just sample data for the list-jobs command. In a real implementation,
	// this would query the JIMM API for actual job data.
	for range 10000 {
		jobs = append(jobs, params.ListJobInfo{
			ID:     int64(len(jobs) + 1),
			Status: "Sample Job",
		})
	}
}

// Run implements Command.Run.
func (c *listjobsCommand) Run(ctxt *cmd.Context) error {
	client, err := c.getJIMMAPI()
	if err != nil {
		return err
	}
	defer client.Close()

	// Create debug file in workspace root (truncates if exists)
	debugFile, err := os.Create("listjobs-debug.log")
	if err == nil {
		defer debugFile.Close()
		fmt.Fprintf(ctxt.Stderr, "Debug log: listjobs-debug.log\n")
	}

	batchNum := 0
	for {
		// resp, err := client.ListJobs(&params.ListJobsRequest{Count: 10})
		// if err != nil {
		// 	return err
		// }

		// Write to debug file BEFORE fetching (line 101)
		if debugFile != nil {
			batchNum++
			fmt.Fprintf(debugFile, "BATCH %d: Fetching at %s\n", batchNum, time.Now().Format("15:04:05.000"))
			debugFile.Sync() // Force write to disk immediately
		}
		resp := getJobs()

		err = c.out.Write(ctxt, resp.Jobs)
		if err != nil {
			return err
		}
		if resp.NextCursor == "" {
			break
		}
	}
	return nil
}

func getJobs() params.ListJobsResponse {
	// time.Sleep(2 * time.Second) // Simulate delay for fetching jobs
	if i >= len(jobs) {
		return params.ListJobsResponse{
			Jobs:       []params.ListJobInfo{},
			NextCursor: "",
		}
	}
	nextIndex := i + 100
	if nextIndex > len(jobs) {
		nextIndex = len(jobs)
	}
	resp := params.ListJobsResponse{
		Jobs:       jobs[i:nextIndex],
		NextCursor: "",
	}
	i = nextIndex
	if i < len(jobs) {
		resp.NextCursor = "cursor" // In a real implementation, this would be a real cursor value.
	}

	return resp
}
