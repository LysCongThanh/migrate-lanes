package detect

import (
	"strconv"
	"strings"
	"time"
)

// timestampLayout is the 14-digit UTC timestamp used by dbmate, Rails,
// Prisma, goose (-s off) and Laravel (with separators removed).
const timestampLayout = "20060102150405"

// trimZeros drops leading zeros so "0005" and "5" compare as one version.
func trimZeros(s string) string {
	t := strings.TrimLeft(s, "0")
	if t == "" && s != "" {
		return "0"
	}
	return t
}

// cmpDigits compares two zero-trimmed decimal strings of any length.
func cmpDigits(a, b string) int {
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	return strings.Compare(a, b)
}

func maxDigits(a, b string) string {
	if a == "" || cmpDigits(b, a) > 0 {
		return b
	}
	return a
}

// allocator hands out versions above every version seen in a directory, so
// suggested renames never collide with each other or with existing files.
type allocator struct {
	last string // zero-trimmed digits of the highest version handed out/seen
}

// next returns the next free version formatted like orig (the raw digits of
// the migration being renamed): timestamps advance by one second, sequences
// by one and keep orig's zero padding.
func (a *allocator) next(orig string) string {
	if len(orig) == len(timestampLayout) {
		if t, err := time.Parse(timestampLayout, a.lastPadded(len(orig))); err == nil {
			out := t.Add(time.Second).Format(timestampLayout)
			a.last = trimZeros(out)
			return out
		}
	}
	n, err := strconv.ParseUint(a.last, 10, 64)
	if err != nil {
		n = 0
	}
	a.last = strconv.FormatUint(n+1, 10)
	return a.lastPadded(len(orig))
}

// lastPadded left-pads the last version to width with zeros.
func (a *allocator) lastPadded(width int) string {
	if len(a.last) >= width {
		return a.last
	}
	return strings.Repeat("0", width-len(a.last)) + a.last
}
