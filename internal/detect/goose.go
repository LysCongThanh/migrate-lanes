package detect

import "regexp"

// NewGoose detects pressly/goose SQL or Go migrations ("<version>_<name>.sql"
// with "-- +goose Up", or "<version>_<name>.go" registering via goose).
// goose rejects duplicate versions and, without -allow-missing, refuses
// migrations older than the current database version.
func NewGoose() Detector {
	return &versionedDetector{
		tool:    "goose",
		pattern: regexp.MustCompile(`^(\d+)(_.+)\.(?:sql|go)$`),
		marker:  regexp.MustCompile(`(?i)\+goose\s+up|goose\.Add(?:Named)?Migration`),
		advice:  "renumber the lane's migration (or `goose fix` after merge)",
	}
}
