// Package renumber turns detector rename suggestions into a plan and, only
// when explicitly asked, applies it inside worktrees.
package renumber

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/detect"
	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// Step is one rename resolved against a lane.
type Step struct {
	detect.Rename
	// Applicable is false for lanes without a worktree (branch-only): the
	// user must check the branch out first.
	Applicable bool   `json:"applicable"`
	Reason     string `json:"reason,omitempty"`
}

// Plan collects every rename suggested by findings, optionally limited to
// one lane. Findings without renames (drizzle, alembic) are returned as
// advice-only so the caller can print their guidance.
func Plan(findings []detect.Finding, lookup func(string) *lane.Lane, only string) (steps []Step, adviceOnly []detect.Finding) {
	seen := map[string]bool{}
	for _, f := range findings {
		if len(f.Renames) == 0 {
			if only == "" || slices.Contains(f.Lanes(), only) {
				adviceOnly = append(adviceOnly, f)
			}
			continue
		}
		for _, r := range f.Renames {
			key := r.Lane + "\x00" + r.From
			if seen[key] || (only != "" && r.Lane != only) {
				continue
			}
			seen[key] = true
			s := Step{Rename: r, Applicable: true}
			if l := lookup(r.Lane); l == nil || l.Path == "" {
				s.Applicable, s.Reason = false, "branch has no worktree; check it out (git worktree add) to apply"
			}
			steps = append(steps, s)
		}
	}
	return steps, adviceOnly
}

// Apply renames files inside each step's worktree. It refuses to overwrite,
// refuses paths escaping the worktree and stops at the first error.
func Apply(steps []Step, lookup func(string) *lane.Lane) (applied int, err error) {
	for _, s := range steps {
		if !s.Applicable {
			continue
		}
		root := lookup(s.Lane).Path
		from, err1 := inside(root, s.From)
		to, err2 := inside(root, s.To)
		if err := errors.Join(err1, err2); err != nil {
			return applied, err
		}
		if _, err := os.Lstat(to); err == nil {
			return applied, fmt.Errorf("%s: refusing to overwrite %s", s.Lane, s.To)
		}
		if err := os.Rename(from, to); err != nil {
			return applied, fmt.Errorf("%s: %w", s.Lane, err)
		}
		applied++
	}
	return applied, nil
}

// inside joins a repo-relative path to root and rejects escapes.
func inside(root, rel string) (string, error) {
	p := filepath.Join(root, filepath.FromSlash(rel))
	r, err := filepath.Rel(root, p)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q escapes worktree %s", rel, root)
	}
	return p, nil
}
