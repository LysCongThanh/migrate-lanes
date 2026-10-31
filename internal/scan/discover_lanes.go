// Package scan discovers lanes (worktrees + branches ahead of base), picks
// migration directories and runs detectors. It is the single entry point the
// CLI commands share.
package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/config"
	"github.com/LysCongThanh/migrate-lanes/internal/gitx"
	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// resolveBase picks the base ref: flag, then config, then main, then master.
func resolveBase(root, flagBase string, cfg config.Config) (string, error) {
	for _, candidate := range []string{flagBase, cfg.Base} {
		if candidate == "" {
			continue
		}
		if !gitx.RefExists(root, candidate) {
			return "", fmt.Errorf("base ref %q not found", candidate)
		}
		return candidate, nil
	}
	for _, guess := range []string{"main", "master"} {
		if gitx.RefExists(root, guess) {
			return guess, nil
		}
	}
	return "", fmt.Errorf("no base branch: pass --base or set base in %s", config.FileName)
}

// discoverLanes returns worktree lanes first (read from disk, uncommitted
// files included), then branches without a worktree that are ahead of base.
func discoverLanes(root, base string, cfg config.Config, remotes bool) ([]*lane.Lane, error) {
	wts, err := gitx.ListWorktrees(root)
	if err != nil {
		return nil, err
	}
	var lanes []*lane.Lane
	checkedOut := map[string]bool{}
	for _, wt := range wts {
		if wt.Bare || wt.Prunable || cfg.Ignored(wt.Branch) {
			continue
		}
		if _, err := os.Stat(wt.Path); err != nil {
			continue
		}
		checkedOut[wt.Branch] = true
		name := "wt:" + wt.Branch
		if wt.Branch == "" {
			name = "wt:" + filepath.Base(wt.Path) + "(detached)"
		}
		l, err := lane.FromWorktree(wt.Path, name, wt.Branch)
		if err != nil {
			return nil, err
		}
		lanes = append(lanes, l)
	}
	localRefs, err := gitx.ListRefs(root, "refs/heads")
	if err != nil {
		return nil, err
	}
	var remoteRefs []string
	if remotes {
		if remoteRefs, err = gitx.ListRefs(root, "refs/remotes"); err != nil {
			return nil, err
		}
	}
	local := map[string]bool{}
	for _, r := range localRefs {
		local[r] = true
	}
	baseShort := base
	if !local[base] {
		_, baseShort, _ = strings.Cut(base, "/") // "origin/main" -> "main"
	}
	for i, ref := range append(localRefs, remoteRefs...) {
		kind, short := lane.KindBranch, ref
		if i >= len(localRefs) {
			kind = lane.KindRemote
			_, short, _ = strings.Cut(ref, "/")
			if local[short] {
				continue // the local branch already represents this lane
			}
		}
		if ref == base || short == baseShort || checkedOut[short] || cfg.Ignored(short) {
			continue
		}
		if n, err := gitx.AheadCount(root, base, ref); err != nil || n == 0 {
			continue
		}
		l, err := lane.FromRef(root, ref, kind, ref)
		if err != nil {
			return nil, err
		}
		lanes = append(lanes, l)
	}
	return lanes, nil
}
