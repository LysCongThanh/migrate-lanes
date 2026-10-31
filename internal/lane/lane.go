// Package lane models one line of parallel work: the base branch, a git
// worktree (read from disk, so uncommitted files count) or a branch without a
// worktree (read from git objects).
package lane

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/gitx"
)

// Kinds of lanes.
const (
	KindBase     = "base"
	KindWorktree = "worktree"
	KindBranch   = "branch"
	KindRemote   = "remote"
	KindMemory   = "memory" // tests only
)

// Lane is a snapshot of repository files for one line of work.
type Lane struct {
	Name  string   `json:"name"`
	Kind  string   `json:"kind"`
	Ref   string   `json:"ref,omitempty"`
	Path  string   `json:"path,omitempty"` // worktree root; empty for branches
	Files []string `json:"-"`              // repo-relative, slash separated, sorted
	read  func(p string) ([]byte, error)
}

// Read returns the content of a repo-relative file in this lane.
func (l *Lane) Read(p string) ([]byte, error) { return l.read(p) }

// FilesUnder returns paths (relative to dir) of every file below dir.
func (l *Lane) FilesUnder(dir string) []string {
	prefix := strings.TrimSuffix(dir, "/") + "/"
	if dir == "" || dir == "." {
		prefix = ""
	}
	var out []string
	for _, f := range l.Files {
		if strings.HasPrefix(f, prefix) {
			out = append(out, strings.TrimPrefix(f, prefix))
		}
	}
	return out
}

// Has reports whether the lane contains the repo-relative file p.
func (l *Lane) Has(p string) bool {
	i := sort.SearchStrings(l.Files, p)
	return i < len(l.Files) && l.Files[i] == p
}

// FromRef builds a lane from committed files at ref.
func FromRef(repo, name, kind, ref string) (*Lane, error) {
	files, err := gitx.TreeFiles(repo, ref)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return &Lane{Name: name, Kind: kind, Ref: ref, Files: files,
		read: func(p string) ([]byte, error) { return gitx.ShowFile(repo, ref, p) }}, nil
}

// FromWorktree builds a lane from the files currently on disk in a worktree,
// including untracked files and excluding deleted ones.
func FromWorktree(root, name, branch string) (*Lane, error) {
	listed, err := gitx.WorkingFiles(root)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(listed))
	seen := map[string]bool{}
	for _, f := range listed {
		if seen[f] {
			continue
		}
		seen[f] = true
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(f))); err == nil {
			files = append(files, f)
		}
	}
	sort.Strings(files)
	return &Lane{Name: name, Kind: KindWorktree, Ref: branch, Path: root, Files: files,
		read: func(p string) ([]byte, error) { return os.ReadFile(filepath.Join(root, filepath.FromSlash(p))) }}, nil
}

// FromMemory builds a lane from an in-memory file map (unit tests).
func FromMemory(name string, files map[string]string) *Lane {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, path.Clean(k))
	}
	sort.Strings(keys)
	return &Lane{Name: name, Kind: KindMemory, Files: keys,
		read: func(p string) ([]byte, error) {
			c, ok := files[p]
			if !ok {
				return nil, os.ErrNotExist
			}
			return []byte(c), nil
		}}
}
