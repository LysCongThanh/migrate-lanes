package detect

import "testing"

func TestGoose(t *testing.T) {
	const d = "sql/schema/"
	up := "-- +goose Up\nCREATE TABLE t();\n-- +goose Down\n"
	base := map[string]string{d + "00001_init.sql": up}
	cases := []struct {
		name  string
		base  map[string]string
		lanes laneFiles
		want  want
	}{
		{"clean", base, laneFiles{"a": with(base, map[string]string{d + "00002_a.sql": up})}, want{}},
		{"sql and go migrations share a version", base, laneFiles{
			"a": with(base, map[string]string{d + "00002_a.sql": up}),
			"b": with(base, map[string]string{d + "00002_b.go": "goose.AddMigrationContext(up, down)"}),
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "00002_b.go": d + "00003_b.go"}}},
		{"no goose marker: not claimed", map[string]string{d + "00001_init.sql": "create table t();"}, laneFiles{
			"a": {d + "00001_init.sql": "", d + "00002_a.sql": ""}, "b": {d + "00001_init.sql": "", d + "00002_b.sql": ""},
		}, want{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewGoose(), tc.base, tc.lanes), tc.want)
		})
	}
}
