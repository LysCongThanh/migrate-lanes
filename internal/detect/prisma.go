package detect

import "regexp"

// NewPrisma detects Prisma Migrate folders:
// "prisma/migrations/<YYYYMMDDHHMMSS>_<name>/migration.sql". Prisma applies
// folders in name order and flags history drift when an older folder appears
// after newer ones were applied.
func NewPrisma() Detector {
	return &versionedDetector{
		tool:       "prisma",
		pattern:    regexp.MustCompile(`^(\d{14})(_.+)$`),
		folderFile: "migration.sql",
		advice:     "rename the lane's migration folder to a newer timestamp, or rebase and re-run `prisma migrate dev`",
	}
}
