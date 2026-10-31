package detect

import "testing"

func TestSqlx(t *testing.T) {
	const d = "migrations/"
	base := map[string]string{d + "20240101000000_init.sql": ""}
	cases := []struct {
		name  string
		base  map[string]string
		lanes laneFiles
		want  want
	}{
		{"clean", base, laneFiles{"a": with(base, map[string]string{d + "20240102000000_a.sql": ""})}, want{}},
		{"same timestamp in two lanes", base, laneFiles{
			"a": with(base, map[string]string{d + "20240102000000_a.sql": ""}),
			"b": with(base, map[string]string{d + "20240102000000_b.sql": ""}),
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "20240102000000_b.sql": d + "20240102000001_b.sql"}}},
		{"top-level files are ignored", map[string]string{"1_notes.sql": ""}, laneFiles{"a": {"1_notes.sql": "", "1_other.sql": ""}}, want{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewSqlx(), tc.base, tc.lanes), tc.want)
		})
	}
}
