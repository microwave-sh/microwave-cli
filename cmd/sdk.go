package cmd

import (
	"context"
	"fmt"

	"github.com/microwave-sh/microwave-cli/internal/output"
)

// SDKCmd is the parent command group for SDK generation operations.
type SDKCmd struct {
	Targets  SDKTargetsCmd  `cmd:"" help:"Manage SDK targets."`
	Runs     SDKRunsCmd     `cmd:"" help:"View and trigger SDK generation runs."`
	Generate SDKGenerateCmd `cmd:"" help:"Trigger a new SDK generation run."`
}

// ── sdk targets ──────────────────────────────────────────────────────────────

// SDKTargetsCmd groups SDK target subcommands.
type SDKTargetsCmd struct {
	List sdkTargetsListCmd `cmd:"" help:"List SDK targets for a connection."`
}

type sdkTargetsListCmd struct {
	Connection string `name:"connection" short:"c" help:"Connection ID." required:""`
}

func (c *sdkTargetsListCmd) Run(ctx context.Context, g *Globals) error {
	targets, err := g.Client().ListSDKTargets(ctx, c.Connection)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(targets)
	}
	rows := make([][]string, len(targets))
	for i, t := range targets {
		rows[i] = []string{t.ID, t.Language, t.Repo}
	}
	output.PrintTable([]string{"ID", "Language", "Repo"}, rows, false)
	return nil
}

// ── sdk runs ─────────────────────────────────────────────────────────────────

// SDKRunsCmd groups SDK run subcommands.
type SDKRunsCmd struct {
	List sdkRunsListCmd `cmd:"" help:"List generation runs for a target."`
}

type sdkRunsListCmd struct {
	Target string `name:"target" short:"t" help:"SDK target ID." required:""`
}

func (c *sdkRunsListCmd) Run(ctx context.Context, g *Globals) error {
	runs, err := g.Client().ListSDKRuns(ctx, c.Target)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(runs)
	}
	rows := make([][]string, len(runs))
	for i, r := range runs {
		rows[i] = []string{r.ID, r.RunType, r.Status, r.PRUrl}
	}
	output.PrintTable([]string{"ID", "Run Type", "Status", "PR URL"}, rows, false)
	return nil
}

// ── sdk generate ──────────────────────────────────────────────────────────────

// SDKGenerateCmd enqueues a new generation run for an SDK target.
type SDKGenerateCmd struct {
	Target string `name:"target" short:"t" help:"SDK target ID." required:""`
}

func (c *SDKGenerateCmd) Run(ctx context.Context, g *Globals) error {
	run, err := g.Client().TriggerRun(ctx, c.Target)
	if err != nil {
		return err
	}
	if g.IsJSON() {
		return output.PrintJSON(run)
	}
	fmt.Printf("Queued run %s (%s).\n", run.ID, run.Status)
	return nil
}
