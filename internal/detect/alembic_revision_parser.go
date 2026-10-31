package detect

import (
	"regexp"
	"strings"
)

// Matches both classic and typed Alembic headers:
//
//	revision = 'abc123'
//	revision: str = "abc123"
//	down_revision: Union[str, None] = ('a', 'b')
var (
	alembicRevRe  = regexp.MustCompile(`(?m)^revision\s*(?::[^=\n]+)?=\s*['"]([^'"]+)['"]`)
	alembicDownRe = regexp.MustCompile(`(?m)^down_revision\s*(?::[^=\n]+)?=\s*(.+)$`)
	quotedRe      = regexp.MustCompile(`['"]([^'"]+)['"]`)
)

// parseAlembicRevision extracts the revision id and parent ids from a
// migration script. ok is false when the file is not an Alembic revision.
func parseAlembicRevision(src []byte) (rev string, downs []string, ok bool) {
	m := alembicRevRe.FindSubmatch(src)
	d := alembicDownRe.FindSubmatch(src)
	if m == nil || d == nil {
		return "", nil, false
	}
	value := strings.TrimSpace(string(d[1]))
	for _, q := range quotedRe.FindAllStringSubmatch(value, -1) {
		downs = append(downs, q[1])
	}
	return string(m[1]), downs, true
}
