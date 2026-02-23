// Copyright 2026 Canonical.

package cmd

import (
	"fmt"
	"strings"

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

The command supports filtering by job kinds and statuses, and allows you to
limit the number of results returned (up to 10,000 jobs).

Valid job statuses are: running, successful, pending, failed, unknown
`
	listjobsCommandExample = `
    juju jobs
    juju jobs --format json
    juju jobs --count 500
    juju jobs --kinds bootstrap-controller,upgrade-to
    juju jobs --statuses running,pending
    juju jobs --count 1000 --statuses failed --kinds bootstrap-controller,upgrade-to
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
	out      cmd.Output
	count    int
	kinds    string
	statuses string
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
	f.IntVar(&c.count, "count", 100, "Maximum number of jobs to return (max 10000)")
	f.StringVar(&c.kinds, "kinds", "", "Filter jobs by kinds (comma-separated)")
	f.StringVar(&c.statuses, "statuses", "", "Filter jobs by statuses (comma-separated)")
}

// Run implements Command.Run.
func (c *listjobsCommand) Run(ctxt *cmd.Context) error {
	if c.count <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}
	if c.count > 10000 {
		return fmt.Errorf("count cannot exceed 10000, got %d", c.count)
	}

	client, err := c.getJIMMAPI()
	if err != nil {
		return err
	}
	defer client.Close()

	var statuses []params.JobStatus
	if c.statuses != "" {
		validStatuses := map[string]params.JobStatus{
			"running":    params.StatusRunning,
			"successful": params.StatusSuccessful,
			"pending":    params.StatusPending,
			"failed":     params.StatusFailed,
			"unknown":    params.StatusUnknown,
		}

		statusStrs := strings.SplitSeq(c.statuses, ",")
		for s := range statusStrs {
			status, ok := validStatuses[s]
			if !ok {
				return fmt.Errorf("invalid status %q, must be one of: running, successful, pending, failed, unknown", s)
			}
			statuses = append(statuses, status)
		}
	}
	kinds := []string{}
	if c.kinds != "" {
		kinds = strings.Split(c.kinds, ",")
	}

	resp, err := client.ListJobs(&params.ListJobsRequest{
		Count:    c.count,
		Kinds:    kinds,
		Statuses: statuses,
	})
	if err != nil {
		return err
	}

	err = c.out.Write(ctxt, resp.Jobs)
	if err != nil {
		return err
	}

	return nil
}
