package detect

import (
	"fmt"
	"strings"
	"testing"
)

// journal builds a drizzle meta/_journal.json with the given tags as idx 0..n.
func journal(tags ...string) string {
	var parts []string
	for i, tag := range tags {
		parts = append(parts, fmt.Sprintf(`{"idx":%d,"version":"7","when":%d,"tag":%q,"breakpoints":true}`, i, 1700000000000+i, tag))
	}
	return `{"version":"7","dialect":"postgresql","entries":[` + strings.Join(parts, ",") + `]}`
}

func TestDrizzle(t *testing.T) {
	const j = "drizzle/meta/_journal.json"
	base := map[string]string{j: journal("0000_init")}
	cases := []struct {
		name  string
		base  map[string]string
		lanes laneFiles
		want  want
	}{
		{"clean", base, laneFiles{"a": {j: journal("0000_init", "0001_red_hulk")}}, want{}},
		{"two lanes generate idx 1", base, laneFiles{
			"a": {j: journal("0000_init", "0001_red_hulk")},
			"b": {j: journal("0000_init", "0001_blue_wolf")},
		}, want{kinds: []string{KindDuplicate}}},
		{"lane generated before base moved", map[string]string{j: journal("0000_init", "0001_merged")}, laneFiles{
			"a": {j: journal("0000_init", "0001_stale")},
		}, want{kinds: []string{KindJournal}}},
		{"conflict markers in journal", base, laneFiles{"a": {j: "<<<<<<< HEAD\n{}"}}, want{kinds: []string{KindParseError}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewDrizzle(), tc.base, tc.lanes), tc.want)
		})
	}
}
