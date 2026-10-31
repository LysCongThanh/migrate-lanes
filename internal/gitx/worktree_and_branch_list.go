package gitx

import (
	"strings"
)

// Worktree is one entry of `git worktree list --porcelain`.
type Worktree struct {
	Path   string
	Head   string
	Branch string // short name, empty when detached
	Bare   bool
	// Prunable is set when git reports the worktree directory is gone.
	Prunable bool
}

// ListWorktrees parses `git worktree list --porcelain` for the repo at dir.
func ListWorktrees(dir string) ([]Worktree, error) {
	out, err := Run(dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreePorcelain(out), nil
}

// parseWorktreePorcelain turns porcelain blocks (blank-line separated) into
// Worktree values. Unknown attribute lines are ignored for forward compat.
func parseWorktreePorcelain(out string) []Worktree {
	var list []Worktree
	var cur *Worktree
	for _, line := range strings.Split(out, "\n") {
		key, val, _ := strings.Cut(strings.TrimSpace(line), " ")
		switch key {
		case "worktree":
			list = append(list, Worktree{Path: val})
			cur = &list[len(list)-1]
		case "HEAD":
			if cur != nil {
				cur.Head = val
			}
		case "branch":
			if cur != nil {
				cur.Branch = strings.TrimPrefix(val, "refs/heads/")
			}
		case "bare":
			if cur != nil {
				cur.Bare = true
			}
		case "prunable":
			if cur != nil {
				cur.Prunable = true
			}
		}
	}
	return list
}

// ListRefs returns short names of refs under the given namespaces, e.g.
// "refs/heads" or "refs/remotes". Symbolic refs like origin/HEAD are skipped.
func ListRefs(dir string, namespaces ...string) ([]string, error) {
	args := append([]string{"for-each-ref", "--format=%(refname:short)%00%(symref)"}, namespaces...)
	out, err := Run(dir, args...)
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, line := range strings.Split(out, "\n") {
		name, symref, _ := strings.Cut(line, "\x00")
		if name == "" || symref != "" || strings.HasSuffix(name, "/HEAD") {
			continue
		}
		refs = append(refs, name)
	}
	return refs, nil
}
