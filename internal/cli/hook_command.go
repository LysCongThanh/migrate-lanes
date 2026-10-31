package cli

import (
	"fmt"
)

// claudeHookSnippet is a Claude Code settings.json fragment. PostToolUse runs
// after Write/Edit/MultiEdit; `check --hook` reads the hook JSON from stdin,
// and on a collision exits 2 so Claude Code shows stderr to the agent (the
// file is already written, so the agent is told to renumber it).
const claudeHookSnippet = `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "migrate-lanes check --hook",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
`

// runHook implements `migrate-lanes hook claude`.
func runHook(args []string, stdio IO) int {
	if len(args) != 1 || args[0] != "claude" {
		fmt.Fprintln(stdio.Err, "usage: migrate-lanes hook claude")
		return ExitUsage
	}
	fmt.Fprint(stdio.Out, claudeHookSnippet)
	fmt.Fprintln(stdio.Err, "# merge into .claude/settings.json (project) or ~/.claude/settings.json (all projects)")
	return ExitClean
}
