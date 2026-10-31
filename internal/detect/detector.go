// Package detect holds one detector per migration tool. A detector claims the
// directories its tool owns and analyzes them across lanes for collisions.
//
// Adding a tool = one new file with a constructor registered in All() plus a
// table test. See CONTRIBUTING.md.
package detect

import (
	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// Detector finds migration collisions for one tool.
type Detector interface {
	// Tool is the stable identifier used in config and output.
	Tool() string
	// Claim returns the migration directories (repo-relative) this tool owns
	// in the given lane.
	Claim(l *lane.Lane) []string
	// Analyze compares lanes against base for a claimed directory.
	Analyze(dir string, base *lane.Lane, lanes []*lane.Lane) []Finding
}

// Finding kinds.
const (
	KindDuplicate  = "duplicate-version"
	KindOutOfOrder = "out-of-order"
	KindJournal    = "journal-divergence"
	KindHeads      = "multiple-heads"
	KindParseError = "parse-error"
)

// Entry points at one migration involved in a finding.
type Entry struct {
	Lane string `json:"lane"`
	File string `json:"file"` // repo-relative
}

// Rename is a suggested fix: move From to To inside Lane's worktree.
type Rename struct {
	Lane string `json:"lane"`
	From string `json:"from"` // repo-relative
	To   string `json:"to"`
}

// Finding is one collision (or problem) in a migration directory.
type Finding struct {
	Tool    string   `json:"tool"`
	Dir     string   `json:"dir"`
	Kind    string   `json:"kind"`
	Version string   `json:"version,omitempty"`
	Message string   `json:"message"`
	Entries []Entry  `json:"entries"`
	Renames []Rename `json:"renames,omitempty"`
	Advice  string   `json:"advice,omitempty"`
}

// Lanes returns the distinct lane names involved in the finding.
func (f Finding) Lanes() []string {
	var out []string
	seen := map[string]bool{}
	for _, e := range f.Entries {
		if !seen[e.Lane] {
			seen[e.Lane] = true
			out = append(out, e.Lane)
		}
	}
	return out
}

// All returns every detector in claim priority order: specific layouts
// (journal files, revision graphs, marker comments) before generic numeric
// filename patterns, so a directory is claimed by the most precise tool.
func All() []Detector {
	return []Detector{
		NewDrizzle(),
		NewAlembic(),
		NewPrisma(),
		NewLaravel(),
		NewRails(),
		NewDbmate(),
		NewGoose(),
		NewGolangMigrate(),
		NewSqlx(),
	}
}

// ByTool returns the detector for a tool name, or nil.
func ByTool(tool string) Detector {
	for _, d := range All() {
		if d.Tool() == tool {
			return d
		}
	}
	return nil
}

// Tools lists every supported tool identifier.
func Tools() []string {
	var out []string
	for _, d := range All() {
		out = append(out, d.Tool())
	}
	return out
}
