package detect

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// Analyze finds duplicate versions (two different migrations claiming the
// same version, across lanes or against base) and out-of-order additions (a
// lane adds a version older than the newest migration already on base).
// Every finding carries rename suggestions from a shared allocator so fixes
// never collide with each other.
func (d *versionedDetector) Analyze(dir string, base *lane.Lane, lanes []*lane.Lane) []Finding {
	baseMigs := d.collect(base, dir)
	// owners: migration key -> lanes holding it, in lane order (base first).
	owners := map[string][]string{}
	byKey := map[string]*migration{}
	for _, m := range baseMigs {
		owners[m.key] = append(owners[m.key], base.Name)
		byKey[m.key] = m
	}
	for _, l := range lanes {
		for _, m := range d.collect(l, dir) {
			owners[m.key] = append(owners[m.key], l.Name)
			if _, ok := byKey[m.key]; !ok {
				byKey[m.key] = m
			}
		}
	}
	baseMax, maxAll := "", ""
	byVersion := map[string][]string{}
	for key, m := range byKey {
		byVersion[m.digits] = append(byVersion[m.digits], key)
		maxAll = maxDigits(maxAll, m.digits)
		if _, ok := baseMigs[key]; ok {
			baseMax = maxDigits(baseMax, m.digits)
		}
	}
	alloc := &allocator{last: maxAll}
	var findings []Finding
	renamed := map[string]bool{}
	for _, v := range sortedKeys(byVersion) {
		keys := byVersion[v]
		if len(keys) < 2 || allInBase(keys, baseMigs) {
			continue
		}
		sort.Slice(keys, func(i, j int) bool {
			return d.rank(keys[i], baseMigs, owners, base, lanes) < d.rank(keys[j], baseMigs, owners, base, lanes)
		})
		f := Finding{Tool: d.tool, Dir: dir, Kind: KindDuplicate, Version: byKey[keys[0]].raw,
			Message: fmt.Sprintf("version %s is used by %d different migrations", byKey[keys[0]].raw, len(keys)), Advice: d.advice}
		for i, key := range keys {
			f.Entries = append(f.Entries, d.entries(dir, byKey[key], owners[key])...)
			if i > 0 {
				f.Renames = append(f.Renames, d.renames(dir, byKey[key], owners[key], base.Name, alloc)...)
				renamed[key] = true
			}
		}
		findings = append(findings, f)
	}
	for _, key := range sortedKeys(byKey) {
		m := byKey[key]
		if _, inBase := baseMigs[key]; inBase || renamed[key] || baseMax == "" || cmpDigits(m.digits, baseMax) >= 0 {
			continue
		}
		findings = append(findings, Finding{Tool: d.tool, Dir: dir, Kind: KindOutOfOrder, Version: m.raw,
			Message: fmt.Sprintf("%s is older than the newest base migration (%s); databases already migrated on base will skip or reject it", m.key, d.prefix(baseMax)),
			Entries: d.entries(dir, m, owners[key]), Renames: d.renames(dir, m, owners[key], base.Name, alloc), Advice: d.advice})
	}
	return findings
}

// rank orders migrations of one version: base first, then by the earliest
// lane that holds it. The first one keeps its version.
func (d *versionedDetector) rank(key string, baseMigs map[string]*migration, owners map[string][]string, base *lane.Lane, lanes []*lane.Lane) int {
	if _, ok := baseMigs[key]; ok {
		return -1
	}
	for i, l := range lanes {
		if contains(owners[key], l.Name) {
			return i
		}
	}
	return len(lanes)
}

func (d *versionedDetector) entries(dir string, m *migration, laneNames []string) []Entry {
	var out []Entry
	for _, ln := range laneNames {
		for _, n := range m.names {
			out = append(out, Entry{Lane: ln, File: path.Join(dir, n)})
		}
	}
	return out
}

// renames gives every non-base lane holding m the same new version.
func (d *versionedDetector) renames(dir string, m *migration, laneNames []string, baseName string, a *allocator) []Rename {
	newPrefix := d.prefix(a.next(onlyDigits(m.raw)))
	var out []Rename
	for _, ln := range laneNames {
		if ln == baseName {
			continue
		}
		for _, n := range m.names {
			out = append(out, Rename{Lane: ln, From: path.Join(dir, n),
				To: path.Join(dir, newPrefix+strings.TrimPrefix(n, m.raw))})
		}
	}
	return out
}

func allInBase(keys []string, baseMigs map[string]*migration) bool {
	for _, k := range keys {
		if _, ok := baseMigs[k]; !ok {
			return false
		}
	}
	return true
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
