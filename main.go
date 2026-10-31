// Command migrate-lanes detects database migration collisions across git
// worktrees and branches before they are merged.
package main

import (
	"os"

	"github.com/LysCongThanh/migrate-lanes/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], cli.IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}))
}
