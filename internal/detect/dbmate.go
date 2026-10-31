package detect

import "regexp"

// NewDbmate detects dbmate migrations: "<YYYYMMDDHHMMSS>_<name>.sql" holding
// a "-- migrate:up" section.
func NewDbmate() Detector {
	return &versionedDetector{
		tool:    "dbmate",
		pattern: regexp.MustCompile(`^(\d{14})(_.*)?\.sql$`),
		marker:  regexp.MustCompile(`--\s*migrate:up`),
		advice:  "give the lane's migration a newer timestamp so it sorts after base",
	}
}
