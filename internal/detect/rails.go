package detect

import "regexp"

// NewRails detects Rails Active Record migrations: "<YYYYMMDDHHMMSS>_<name>.rb".
// Rails raises DuplicateMigrationVersionError on duplicate timestamps.
func NewRails() Detector {
	return &versionedDetector{
		tool:    "rails",
		pattern: regexp.MustCompile(`^(\d{14})(_[a-z0-9_]+)\.rb$`),
		advice:  "give the lane's migration a newer timestamp (and run db:migrate again)",
	}
}
