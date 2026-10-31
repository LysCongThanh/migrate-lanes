// Package cli implements the migrate-lanes commands on top of package scan.
//
// Exit codes: 0 clean, 1 collisions found, 2 usage/config error. The
// `check --hook` mode follows Claude Code hook semantics instead: exit 2
// with the message on stderr means "collision, tell the agent".
package cli

import (
	"flag"
	"fmt"
	"io"
)

// Exit codes shared by every command.
const (
	ExitClean     = 0
	ExitCollision = 1
	ExitUsage     = 2
)

// Version is set at build time via -ldflags "-X .../internal/cli.Version=...".
var Version = "dev"

// IO bundles the streams so tests can capture output.
type IO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

const usage = `migrate-lanes - catch database migration collisions across git worktrees and branches before merge

Usage:
  migrate-lanes scan     [--base REF] [--remotes] [--json] [-C DIR]
  migrate-lanes check    [--base REF] [--hook] [FILE...]
  migrate-lanes renumber [--base REF] [--lane NAME] [--dry-run | --apply] [--json]
  migrate-lanes hook claude
  migrate-lanes version

Exit codes: 0 clean, 1 collisions found, 2 usage/config error
(check --hook: 2 = collision, message on stderr for the agent).
`

// Main runs the CLI and returns the process exit code.
func Main(args []string, stdio IO) int {
	if len(args) == 0 {
		fmt.Fprint(stdio.Err, usage)
		return ExitUsage
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "scan":
		return runScan(rest, stdio)
	case "check":
		return runCheck(rest, stdio)
	case "renumber":
		return runRenumber(rest, stdio)
	case "hook":
		return runHook(rest, stdio)
	case "version", "--version":
		fmt.Fprintln(stdio.Out, "migrate-lanes", Version)
		return ExitClean
	case "help", "-h", "--help":
		fmt.Fprint(stdio.Out, usage)
		return ExitClean
	default:
		fmt.Fprintf(stdio.Err, "unknown command %q\n\n%s", cmd, usage)
		return ExitUsage
	}
}

// commonFlags are shared by scan, check and renumber.
type commonFlags struct {
	base    string
	dir     string
	remotes bool
	json    bool
}

func newFlagSet(name string, stdio IO, c *commonFlags) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stdio.Err)
	fs.StringVar(&c.base, "base", "", "base branch/ref (default: config, then main, then master)")
	fs.StringVar(&c.dir, "C", ".", "run as if started in this directory")
	fs.BoolVar(&c.remotes, "remotes", false, "also scan remote-tracking branches ahead of base")
	fs.BoolVar(&c.json, "json", false, "machine-readable output")
	return fs
}

// parse wraps FlagSet.Parse, mapping -h to exit 0 and bad flags to exit 2.
func parse(fs *flag.FlagSet, args []string) (code int, ok bool) {
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return ExitClean, false
		}
		return ExitUsage, false
	}
	return 0, true
}
