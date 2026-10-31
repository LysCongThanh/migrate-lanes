package detect

import "testing"

func TestDbmate(t *testing.T) {
	const d = "db/migrations/"
	m := "-- migrate:up\ncreate table t();\n-- migrate:down\n"
	base := map[string]string{d + "20240101120000_init.sql": m, d + "20240301120000_x.sql": m}
	cases := []struct {
		name  string
		base  map[string]string
		lanes laneFiles
		want  want
	}{
		{"newer timestamp is clean", base, laneFiles{"a": with(base, map[string]string{d + "20240401000000_a.sql": m})}, want{}},
		{"older timestamp than base head", base, laneFiles{"a": with(base, map[string]string{d + "20240201000000_a.sql": m})},
			want{kinds: []string{KindOutOfOrder}, renames: map[string]string{d + "20240201000000_a.sql": d + "20240301120001_a.sql"}}},
		{"timestamp rename rolls over the day", map[string]string{d + "20241231235959_y.sql": m}, laneFiles{
			"a": {d + "20241231235959_y.sql": m, d + "20241231235959_z.sql": m},
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "20241231235959_z.sql": d + "20250101000000_z.sql"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewDbmate(), tc.base, tc.lanes), tc.want)
		})
	}
}
