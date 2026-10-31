package detect

import "regexp"

// NewSqlx detects sqlx-cli simple migrations ("<version>_<name>.sql" with no
// tool marker). Reversible sqlx pairs (.up.sql/.down.sql) are claimed by the
// golang-migrate detector first; the rules are identical.
func NewSqlx() Detector {
	return &versionedDetector{
		tool:    "sqlx",
		pattern: regexp.MustCompile(`^(\d+)(_.+?)(?:\.up|\.down)?\.sql$`),
		advice:  "renumber the lane's migration; sqlx rejects duplicate versions",
	}
}
