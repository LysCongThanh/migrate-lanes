package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/detect"
	"github.com/LysCongThanh/migrate-lanes/internal/scan"
)

// hookInput is the subset of the Claude Code hook stdin payload we read.
type hookInput struct {
	Cwd       string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

// runCheck implements `migrate-lanes check`: report only collisions that
// involve the given files (or, with no files, the current worktree).
func runCheck(args []string, stdio IO) int {
	var c commonFlags
	var hook bool
	fs := newFlagSet("check", stdio, &c)
	fs.BoolVar(&hook, "hook", false, "read a Claude Code PostToolUse payload on stdin; exit 2 + stderr on collision")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	files := fs.Args()
	if hook {
		var in hookInput
		if err := json.NewDecoder(io.LimitReader(stdio.In, 4<<20)).Decode(&in); err != nil || in.ToolInput.FilePath == "" {
			return ExitClean // not a file-writing tool call: nothing to check
		}
		p := in.ToolInput.FilePath
		if !filepath.IsAbs(p) {
			p = filepath.Join(in.Cwd, p)
		}
		if !migrationExt[strings.ToLower(filepath.Ext(p))] {
			return ExitClean // fast path: cannot be a migration file
		}
		files = []string{p}
	}
	dir := c.dir
	if len(files) > 0 {
		dir = filepath.Dir(absPath(files[0]))
		if _, err := os.Stat(dir); err != nil {
			return ExitClean
		}
	}
	res, err := scan.Run(scan.Options{Dir: dir, Base: c.base, Remotes: c.remotes})
	if err != nil {
		if hook && errors.Is(err, scan.ErrNotRepo) {
			return ExitClean
		}
		if hook {
			// Non-blocking: never stop the agent because the tool is misconfigured.
			fmt.Fprintln(stdio.Err, "migrate-lanes (hook skipped):", err)
			return ExitCollision
		}
		fmt.Fprintln(stdio.Err, "migrate-lanes:", err)
		return ExitUsage
	}
	hits, current := relevantFindings(res, files)
	if len(hits) == 0 {
		return ExitClean
	}
	w := stdio.Out
	if hook {
		w = stdio.Err
		fmt.Fprintf(w, "migrate-lanes: the migration you just wrote collides with another lane (worktree/branch) of this repo.\n")
	}
	for _, f := range hits {
		printFinding(w, f)
	}
	if hook {
		fmt.Fprintf(w, "\nPick a free version (see fix lines) or run `migrate-lanes renumber --apply --lane %s`, then continue.\n", current)
		return ExitUsage // Claude Code: exit 2 feeds stderr back to the agent
	}
	return ExitCollision
}

// migrationExt lists extensions any supported tool writes.
var migrationExt = map[string]bool{".sql": true, ".py": true, ".rb": true, ".php": true, ".go": true, ".json": true}

// relevantFindings keeps findings that touch the current worktree's lane and
// (when files are given) one of the files' migration directories.
func relevantFindings(res *scan.Result, files []string) (out []detect.Finding, current string) {
	root := realPath(res.Repo)
	for _, l := range res.Lanes {
		if l.Path != "" && realPath(l.Path) == root {
			current = l.Name
		}
	}
	var rels []string
	for _, f := range files {
		rel, err := filepath.Rel(root, realPath(absPath(f)))
		if err == nil && !strings.HasPrefix(rel, "..") {
			rels = append(rels, filepath.ToSlash(rel))
		}
	}
	for _, f := range res.Findings {
		if current == "" || !contains(f.Lanes(), current) {
			continue
		}
		if len(files) > 0 && !anyUnder(rels, f.Dir) {
			continue
		}
		out = append(out, f)
	}
	return out, current
}

func anyUnder(rels []string, dir string) bool {
	for _, r := range rels {
		if strings.HasPrefix(r, dir+"/") {
			return true
		}
	}
	return false
}

func absPath(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}

// realPath resolves symlinks of the longest existing prefix (macOS /tmp).
func realPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	if dir := filepath.Dir(p); dir != p {
		return filepath.Join(realPath(dir), filepath.Base(p))
	}
	return p
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
