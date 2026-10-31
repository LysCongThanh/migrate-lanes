package detect

import (
	"sort"

	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// Target is a migration directory paired with the detector that owns it.
type Target struct {
	Dir      string   `json:"dir"`
	Tool     string   `json:"tool"`
	Detector Detector `json:"-"`
}

// AutoDetect runs every detector's Claim over all lanes. The first detector
// (priority order) to claim a directory wins; a directory nested inside an
// already claimed one is ignored so one migration set is never analyzed twice.
func AutoDetect(lanes []*lane.Lane) []Target {
	owner := map[string]Detector{}
	var dirs []string
	for _, d := range All() {
		for _, l := range lanes {
			for _, dir := range d.Claim(l) {
				if _, taken := owner[dir]; taken || insideAny(dir, dirs) {
					continue
				}
				owner[dir] = d
				dirs = append(dirs, dir)
			}
		}
	}
	sort.Strings(dirs)
	out := make([]Target, 0, len(dirs))
	for _, dir := range dirs {
		out = append(out, Target{Dir: dir, Tool: owner[dir].Tool(), Detector: owner[dir]})
	}
	return out
}

// insideAny reports whether dir is strictly below one of parents.
func insideAny(dir string, parents []string) bool {
	for _, p := range parents {
		if len(dir) > len(p)+1 && dir[:len(p)+1] == p+"/" {
			return true
		}
	}
	return false
}

// ClaimForDir asks every detector (priority order) whether it owns dir in any
// lane. Used when config pins a path but not a tool.
func ClaimForDir(dir string, lanes []*lane.Lane) Detector {
	for _, d := range All() {
		for _, l := range lanes {
			for _, c := range d.Claim(l) {
				if c == dir {
					return d
				}
			}
		}
	}
	return nil
}
