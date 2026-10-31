package detect

import (
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

// versionedDetector covers every tool whose migrations are files (or
// folders) named "<version>_<name>...": numeric sequences (golang-migrate,
// goose, sqlx) and timestamps (dbmate, Rails, Laravel, Prisma). Each tool file
// only fills in this spec.
type versionedDetector struct {
	tool string
	// pattern matches an entry name. Group 1 = raw version prefix, group 2 =
	// the rest of the identity (description). Suffixes outside the groups
	// (".up.sql", ".down.sql") let several files share one migration key.
	pattern *regexp.Regexp
	// folderFile, when set, means entries are folders holding this file
	// (Prisma: "<ts>_<name>/migration.sql").
	folderFile string
	// marker, when set, must match the content of at least one sampled
	// entry file before the directory is claimed (e.g. "-- +goose Up").
	marker *regexp.Regexp
	// formatPrefix turns version digits back into a filename prefix. Nil
	// means the digits are the prefix.
	formatPrefix func(digits string) string
	advice       string
}

func (d *versionedDetector) Tool() string { return d.tool }

// entryOf maps a repo-relative file to (dir, entry name) when it belongs to
// this tool's layout.
func (d *versionedDetector) entryOf(file string) (dir, name string, ok bool) {
	if d.folderFile != "" {
		if path.Base(file) != d.folderFile {
			return "", "", false
		}
		folder := path.Dir(file)
		dir, name = path.Dir(folder), path.Base(folder)
	} else {
		dir, name = path.Dir(file), path.Base(file)
	}
	if dir == "." || !d.pattern.MatchString(name) {
		return "", "", false
	}
	return dir, name, true
}

// Claim groups matching files by directory and keeps directories that pass
// the optional content marker.
func (d *versionedDetector) Claim(l *lane.Lane) []string {
	samples := map[string][]string{}
	for _, f := range l.Files {
		if dir, _, ok := d.entryOf(f); ok {
			samples[dir] = append(samples[dir], f)
		}
	}
	var dirs []string
	for dir, files := range samples {
		if d.marker == nil || d.markerFound(l, files) {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	return dirs
}

// markerFound reads up to three files and reports whether any carries the
// tool's marker comment.
func (d *versionedDetector) markerFound(l *lane.Lane, files []string) bool {
	for i, f := range files {
		if i >= 3 {
			break
		}
		if b, err := l.Read(f); err == nil && d.marker.Match(b) {
			return true
		}
	}
	return false
}

// migration is one logical migration (possibly several files) in a lane.
type migration struct {
	key    string   // version prefix + description, e.g. "0005_add_users"
	raw    string   // raw version prefix as written in the filename
	digits string   // version with separators removed, comparable
	names  []string // entry names (files or folders) directly under dir
}

// collect returns migrations in dir for one lane, keyed by migration key.
func (d *versionedDetector) collect(l *lane.Lane, dir string) map[string]*migration {
	out := map[string]*migration{}
	for _, f := range l.Files {
		fdir, name, ok := d.entryOf(f)
		if !ok || fdir != dir {
			continue
		}
		m := d.pattern.FindStringSubmatch(name)
		key := m[1] + m[2]
		mig, exists := out[key]
		if !exists {
			mig = &migration{key: key, raw: m[1], digits: trimZeros(onlyDigits(m[1]))}
			out[key] = mig
		}
		if !contains(mig.names, name) {
			mig.names = append(mig.names, name)
		}
	}
	return out
}

func (d *versionedDetector) prefix(digits string) string {
	if d.formatPrefix != nil {
		return d.formatPrefix(digits)
	}
	return digits
}

func onlyDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
