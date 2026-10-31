package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/detect"
	"github.com/LysCongThanh/migrate-lanes/internal/scan"
)

// runScan implements `migrate-lanes scan`.
func runScan(args []string, stdio IO) int {
	var c commonFlags
	fs := newFlagSet("scan", stdio, &c)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stdio.Err, "scan: unexpected argument %q\n", fs.Arg(0))
		return ExitUsage
	}
	res, err := scan.Run(scan.Options{Dir: c.dir, Base: c.base, Remotes: c.remotes})
	if err != nil {
		fmt.Fprintln(stdio.Err, "migrate-lanes:", err)
		return ExitUsage
	}
	if c.json {
		writeJSON(stdio.Out, res)
	} else {
		printScan(stdio.Out, res)
	}
	if len(res.Findings) > 0 {
		return ExitCollision
	}
	return ExitClean
}

func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// printScan renders a human summary: lanes, directories, then findings.
func printScan(w io.Writer, res *scan.Result) {
	fmt.Fprintf(w, "base %s · %d lane(s) · %d migration dir(s)\n", res.Base, len(res.Lanes), len(res.Targets))
	for _, t := range res.Targets {
		fmt.Fprintf(w, "  %-15s %s\n", t.Tool, t.Dir)
	}
	if len(res.Targets) == 0 {
		fmt.Fprintln(w, "  no migration directories found (add paths to .migrate-lanes.toml?)")
	}
	if len(res.Findings) == 0 {
		fmt.Fprintln(w, "\nno collisions")
		return
	}
	fmt.Fprintf(w, "\n%d collision(s):\n", len(res.Findings))
	for _, f := range res.Findings {
		printFinding(w, f)
	}
	fmt.Fprintln(w, "\nrun `migrate-lanes renumber` to see suggested fixes")
}

// printFinding renders one finding with its files and suggested fix.
func printFinding(w io.Writer, f detect.Finding) {
	fmt.Fprintf(w, "\n[%s] %s %s: %s\n", f.Kind, f.Tool, f.Dir, f.Message)
	for _, e := range f.Entries {
		fmt.Fprintf(w, "    %-28s %s\n", e.Lane, e.File)
	}
	for _, r := range f.Renames {
		fmt.Fprintf(w, "    fix (%s): %s -> %s\n", r.Lane, r.From, lastSegment(r.To))
	}
	if f.Advice != "" {
		fmt.Fprintf(w, "    hint: %s\n", f.Advice)
	}
}

func lastSegment(p string) string {
	return p[strings.LastIndex(p, "/")+1:]
}
