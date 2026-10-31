package detect

import "regexp"

// NewLaravel detects Laravel migrations: "YYYY_MM_DD_HHMMSS_<name>.php".
// Laravel runs pending files in filename order, so an older timestamp merged
// late runs before newer, already-applied schema changes on fresh databases.
func NewLaravel() Detector {
	return &versionedDetector{
		tool:         "laravel",
		pattern:      regexp.MustCompile(`^(\d{4}_\d{2}_\d{2}_\d{6})(_.+)\.php$`),
		formatPrefix: laravelPrefix,
		advice:       "give the lane's migration a newer timestamp so it sorts after base",
	}
}

// laravelPrefix formats 14 timestamp digits as "YYYY_MM_DD_HHMMSS".
func laravelPrefix(d string) string {
	if len(d) != 14 {
		return d
	}
	return d[0:4] + "_" + d[4:6] + "_" + d[6:8] + "_" + d[8:14]
}
