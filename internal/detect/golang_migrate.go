package detect

import "regexp"

// NewGolangMigrate detects golang-migrate (and sqlx reversible) layouts:
// "<seq>_<name>.up.sql" + "<seq>_<name>.down.sql". golang-migrate refuses
// duplicate versions and silently never applies a version lower than the
// database's current one (golang-migrate/migrate#470).
func NewGolangMigrate() Detector {
	return &versionedDetector{
		tool:    "golang-migrate",
		pattern: regexp.MustCompile(`^(\d+)(_.+)\.(?:up|down)\.sql$`),
		advice:  "rename the lane's .up.sql/.down.sql pair to a free version (migrate-lanes renumber)",
	}
}
