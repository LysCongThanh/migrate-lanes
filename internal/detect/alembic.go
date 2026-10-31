package detect

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// alembicDetector checks Alembic "versions/" folders. Each revision names its
// parent via down_revision; two lanes branching from the same head create a
// second head, and `alembic upgrade head` fails until someone merges them.
type alembicDetector struct{}

// NewAlembic returns the Alembic revision-graph detector.
func NewAlembic() Detector { return alembicDetector{} }

func (alembicDetector) Tool() string { return "alembic" }

func (alembicDetector) Claim(l *lane.Lane) []string {
	samples := map[string][]string{}
	for _, f := range l.Files {
		if strings.HasSuffix(f, ".py") && path.Base(path.Dir(f)) == "versions" && !strings.HasPrefix(path.Base(f), "__") {
			samples[path.Dir(f)] = append(samples[path.Dir(f)], f)
		}
	}
	var dirs []string
	for dir, files := range samples {
		for i, f := range files {
			if i >= 3 {
				break
			}
			if b, err := l.Read(f); err == nil {
				if _, _, ok := parseAlembicRevision(b); ok {
					dirs = append(dirs, dir)
					break
				}
			}
		}
	}
	sort.Strings(dirs)
	return dirs
}

// alembicRev is one revision node in the merged graph of all lanes.
type alembicRev struct {
	downs []string
	file  string
	lanes []string
}

// alembicGraph reads revisions in dir. skip lists files already read from
// base (identical paths are assumed identical revisions).
func alembicGraph(l *lane.Lane, dir string, graph map[string]*alembicRev, skip func(string) bool) []Finding {
	var findings []Finding
	for _, rel := range l.FilesUnder(dir) {
		file := path.Join(dir, rel)
		if strings.Contains(rel, "/") || !strings.HasSuffix(rel, ".py") || skip(file) {
			continue
		}
		b, err := l.Read(file)
		if err != nil {
			continue
		}
		rev, downs, ok := parseAlembicRevision(b)
		if !ok {
			continue
		}
		node, exists := graph[rev]
		if exists && node.file != file {
			findings = append(findings, Finding{Tool: "alembic", Dir: dir, Kind: KindDuplicate, Version: rev,
				Message: fmt.Sprintf("revision id %s is defined by two different files", rev),
				Entries: []Entry{{Lane: node.lanes[0], File: node.file}, {Lane: l.Name, File: file}}})
			continue
		}
		if !exists {
			node = &alembicRev{downs: downs, file: file}
			graph[rev] = node
		}
		node.lanes = append(node.lanes, l.Name)
	}
	return findings
}

// alembicHeads returns revisions nobody points at, sorted.
func alembicHeads(graph map[string]*alembicRev) []string {
	parent := map[string]bool{}
	for _, n := range graph {
		for _, d := range n.downs {
			parent[d] = true
		}
	}
	var heads []string
	for rev := range graph {
		if !parent[rev] {
			heads = append(heads, rev)
		}
	}
	sort.Strings(heads)
	return heads
}

func (alembicDetector) Analyze(dir string, base *lane.Lane, lanes []*lane.Lane) []Finding {
	graph := map[string]*alembicRev{}
	findings := alembicGraph(base, dir, graph, func(string) bool { return false })
	baseHeads := alembicHeads(graph)
	for _, l := range lanes {
		findings = append(findings, alembicGraph(l, dir, graph, base.Has)...)
	}
	heads := alembicHeads(graph)
	if len(heads) <= len(baseHeads) || len(heads) < 2 {
		return findings
	}
	f := Finding{Tool: "alembic", Dir: dir, Kind: KindHeads, Version: strings.Join(heads, ","),
		Message: fmt.Sprintf("merging all lanes leaves %d heads (%s); `alembic upgrade head` will fail", len(heads), strings.Join(heads, ", "))}
	for _, h := range heads {
		for _, ln := range graph[h].lanes {
			f.Entries = append(f.Entries, Entry{Lane: ln, File: graph[h].file})
		}
	}
	f.Advice = fmt.Sprintf("rebase one lane and point its first new revision's down_revision at %q, or run `alembic merge heads` after merging", heads[0])
	return append(findings, f)
}
