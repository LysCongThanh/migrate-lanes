// Package gitx wraps the git CLI. Every call runs `git -C <dir> ...` so the
// tool never depends on the process working directory.
package gitx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes git in dir and returns trimmed stdout. Stderr is folded into
// the error so callers get an actionable message.
func Run(dir string, args ...string) (string, error) {
	out, err := RunRaw(dir, args...)
	return strings.TrimSpace(string(out)), err
}

// RunRaw executes git in dir and returns stdout untouched (needed for -z
// output and file contents).
func RunRaw(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// TopLevel returns the root of the worktree containing dir.
func TopLevel(dir string) (string, error) {
	return Run(dir, "rev-parse", "--show-toplevel")
}

// RefExists reports whether ref resolves to a commit.
func RefExists(dir, ref string) bool {
	_, err := Run(dir, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	return err == nil
}

// AheadCount returns how many commits ref has that base does not.
func AheadCount(dir, base, ref string) (int, error) {
	out, err := Run(dir, "rev-list", "--count", base+".."+ref)
	if err != nil {
		return 0, err
	}
	var n int
	_, err = fmt.Sscanf(out, "%d", &n)
	return n, err
}

// splitNul splits NUL-terminated git output, dropping the trailing empty item.
func splitNul(b []byte) []string {
	var out []string
	for _, p := range strings.Split(string(b), "\x00") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// TreeFiles lists every file path committed at ref (repo-relative, slash form).
func TreeFiles(dir, ref string) ([]string, error) {
	b, err := RunRaw(dir, "ls-tree", "-r", "--name-only", "-z", ref)
	if err != nil {
		return nil, err
	}
	return splitNul(b), nil
}

// WorkingFiles lists tracked plus untracked-but-not-ignored files of a
// worktree. Uncommitted migrations an agent just wrote are included.
func WorkingFiles(worktree string) ([]string, error) {
	b, err := RunRaw(worktree, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	return splitNul(b), nil
}

// ShowFile returns the content of path at ref.
func ShowFile(dir, ref, path string) ([]byte, error) {
	return RunRaw(dir, "cat-file", "blob", ref+":"+path)
}
