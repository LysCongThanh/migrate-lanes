package detect

import "testing"

func TestPrisma(t *testing.T) {
	const d = "prisma/migrations/"
	base := map[string]string{d + "20240101000000_init/migration.sql": "", d + "migration_lock.toml": ""}
	cases := []struct {
		name  string
		lanes laneFiles
		want  want
	}{
		{"clean", laneFiles{"a": with(base, map[string]string{d + "20240102000000_a/migration.sql": ""})}, want{}},
		{"same folder timestamp renames the folder", laneFiles{
			"a": with(base, map[string]string{d + "20240102000000_a/migration.sql": ""}),
			"b": with(base, map[string]string{d + "20240102000000_b/migration.sql": ""}),
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "20240102000000_b": d + "20240102000001_b"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewPrisma(), base, tc.lanes), tc.want)
		})
	}
}
