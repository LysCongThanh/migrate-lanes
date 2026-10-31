package cli

import (
	"fmt"

	"github.com/LysCongThanh/migrate-lanes/internal/detect"
	"github.com/LysCongThanh/migrate-lanes/internal/renumber"
	"github.com/LysCongThanh/migrate-lanes/internal/scan"
)

// runRenumber implements `migrate-lanes renumber`. Dry-run is the default;
// files are only renamed with --apply.
func runRenumber(args []string, stdio IO) int {
	var c commonFlags
	var apply, dryRun bool
	var only string
	fs := newFlagSet("renumber", stdio, &c)
	fs.BoolVar(&apply, "apply", false, "rename files inside worktrees (default: dry run)")
	fs.BoolVar(&dryRun, "dry-run", false, "only print the plan (default)")
	fs.StringVar(&only, "lane", "", "limit fixes to one lane, e.g. wt:feature-a")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if apply && dryRun {
		fmt.Fprintln(stdio.Err, "renumber: --apply and --dry-run are mutually exclusive")
		return ExitUsage
	}
	res, err := scan.Run(scan.Options{Dir: c.dir, Base: c.base, Remotes: c.remotes})
	if err != nil {
		fmt.Fprintln(stdio.Err, "migrate-lanes:", err)
		return ExitUsage
	}
	if only != "" && res.LaneByName(only) == nil {
		fmt.Fprintf(stdio.Err, "renumber: unknown lane %q\n", only)
		return ExitUsage
	}
	steps, advice := renumber.Plan(res.Findings, res.LaneByName, only)
	if c.json {
		writeJSON(stdio.Out, map[string]any{"apply": apply, "steps": steps, "advice_only": advice})
	} else {
		printPlan(stdio, steps, advice, apply)
	}
	if len(steps) == 0 && len(advice) == 0 {
		return ExitClean
	}
	if !apply {
		return ExitCollision
	}
	n, err := renumber.Apply(steps, res.LaneByName)
	if err != nil {
		fmt.Fprintf(stdio.Err, "renumber: applied %d rename(s), then failed: %v\n", n, err)
		return ExitUsage
	}
	if !c.json {
		fmt.Fprintf(stdio.Out, "\napplied %d rename(s); review with `git status` in each worktree and commit\n", n)
	}
	if len(advice) > 0 || n < len(steps) {
		return ExitCollision // something still needs a manual fix
	}
	return ExitClean
}

func printPlan(stdio IO, steps []renumber.Step, advice []detect.Finding, apply bool) {
	w := stdio.Out
	if len(steps) == 0 && len(advice) == 0 {
		fmt.Fprintln(w, "no collisions; nothing to renumber")
		return
	}
	mode := "dry run (pass --apply to rename)"
	if apply {
		mode = "applying"
	}
	fmt.Fprintf(w, "renumber plan, %s:\n", mode)
	for _, s := range steps {
		note := ""
		if !s.Applicable {
			note = "  [skipped: " + s.Reason + "]"
		}
		fmt.Fprintf(w, "  %-24s %s -> %s%s\n", s.Lane, s.From, lastSegment(s.To), note)
	}
	for _, f := range advice {
		fmt.Fprintf(w, "  manual fix [%s] %s %s: %s\n", f.Kind, f.Tool, f.Dir, f.Message)
		if f.Advice != "" {
			fmt.Fprintf(w, "    hint: %s\n", f.Advice)
		}
	}
}
