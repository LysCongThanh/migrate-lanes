package detect

import (
	"sort"
	"testing"

	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// laneFiles is a lane snapshot for table tests: lane name -> path -> content.
type laneFiles map[string]map[string]string

// runDetector claims dirs with d over base + lanes (sorted by name, like a
// stable discovery order) and analyzes each claimed dir.
func runDetector(t *testing.T, d Detector, base map[string]string, lanes laneFiles) []Finding {
	t.Helper()
	b := lane.FromMemory("base", base)
	names := make([]string, 0, len(lanes))
	for n := range lanes {
		names = append(names, n)
	}
	sort.Strings(names)
	var ls []*lane.Lane
	for _, n := range names {
		ls = append(ls, lane.FromMemory(n, lanes[n]))
	}
	dirs := map[string]bool{}
	for _, l := range append([]*lane.Lane{b}, ls...) {
		for _, dir := range d.Claim(l) {
			dirs[dir] = true
		}
	}
	var out []Finding
	for _, dir := range sortedKeys(dirs) {
		out = append(out, d.Analyze(dir, b, ls)...)
	}
	return out
}

// with copies base and adds extra files (a lane forked from base).
func with(base map[string]string, extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// want describes the expected findings of a table case.
type want struct {
	kinds   []string          // finding kinds in order
	renames map[string]string // From -> To (all findings)
}

func assertFindings(t *testing.T, got []Finding, w want) {
	t.Helper()
	var kinds []string
	renames := map[string]string{}
	for _, f := range got {
		kinds = append(kinds, f.Kind)
		for _, r := range f.Renames {
			renames[r.From] = r.To
		}
	}
	if len(kinds) != len(w.kinds) {
		t.Fatalf("kinds = %v, want %v\nfindings: %+v", kinds, w.kinds, got)
	}
	for i := range kinds {
		if kinds[i] != w.kinds[i] {
			t.Fatalf("kinds = %v, want %v", kinds, w.kinds)
		}
	}
	if w.renames == nil {
		return
	}
	if len(renames) != len(w.renames) {
		t.Fatalf("renames = %v, want %v", renames, w.renames)
	}
	for from, to := range w.renames {
		if renames[from] != to {
			t.Fatalf("rename %s = %q, want %q (all: %v)", from, renames[from], to, renames)
		}
	}
}
