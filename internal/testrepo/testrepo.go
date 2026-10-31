// Package testrepo builds throwaway git repositories with real worktrees
// and branches for tests. Nothing is mocked: every helper shells out to git.
package testrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Repo is a temp repository whose main worktree is Dir, on branch main.
type Repo struct {
	t   testing.TB
	Dir string
}

// New creates a repo on branch "main" with base files committed. Git runs
// with an isolated config so the developer's global settings never leak in.
func New(t testing.TB, base map[string]string) *Repo {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.invalid")
	r := &Repo{t: t, Dir: t.TempDir()}
	r.Git(r.Dir, "init", "-q", "-b", "main")
	if base == nil {
		base = map[string]string{}
	}
	base["README.md"] = "test repo\n"
	r.Write(r.Dir, base)
	r.Commit(r.Dir, "base")
	return r
}

// Git runs git in dir and fails the test on error.
func (r *Repo) Git(dir string, args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// Write creates files (repo-relative slash paths) under dir.
func (r *Repo) Write(dir string, files map[string]string) {
	r.t.Helper()
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
}

// Commit stages everything in dir and commits.
func (r *Repo) Commit(dir, msg string) {
	r.t.Helper()
	r.Git(dir, "add", "-A")
	r.Git(dir, "commit", "-q", "--allow-empty", "-m", msg)
}

// Worktree adds a linked worktree on a new branch forked from main and
// writes files into it. commit=false leaves them untracked, like an agent
// that has not committed yet.
func (r *Repo) Worktree(branch string, files map[string]string, commit bool) string {
	r.t.Helper()
	path := filepath.Join(r.t.TempDir(), strings.ReplaceAll(branch, "/", "-"))
	r.Git(r.Dir, "worktree", "add", "-q", "-b", branch, path, "main")
	r.Write(path, files)
	if commit {
		r.Commit(path, "add "+branch)
	}
	return path
}

// Branch creates a branch from main with committed files but no worktree.
func (r *Repo) Branch(branch string, files map[string]string) {
	r.t.Helper()
	path := r.Worktree(branch, files, true)
	r.Git(r.Dir, "worktree", "remove", "--force", path)
}

// AdvanceMain commits files directly on main (simulates a merged PR).
func (r *Repo) AdvanceMain(files map[string]string) {
	r.t.Helper()
	r.Write(r.Dir, files)
	r.Commit(r.Dir, "advance main")
}
