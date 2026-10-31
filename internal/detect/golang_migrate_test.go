package detect

import "testing"

func TestGolangMigrate(t *testing.T) {
	const d = "db/migrations/"
	base := map[string]string{d + "000001_init.up.sql": "", d + "000001_init.down.sql": ""}
	cases := []struct {
		name  string
		base  map[string]string
		lanes laneFiles
		want  want
	}{
		{"clean single lane", base, laneFiles{"a": with(base, map[string]string{d + "000002_a.up.sql": "", d + "000002_a.down.sql": ""})}, want{}},
		{"two lanes same next version", base, laneFiles{
			"a": with(base, map[string]string{d + "000002_users.up.sql": "", d + "000002_users.down.sql": ""}),
			"b": with(base, map[string]string{d + "000002_orders.up.sql": "", d + "000002_orders.down.sql": ""}),
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{
			d + "000002_orders.up.sql": d + "000003_orders.up.sql", d + "000002_orders.down.sql": d + "000003_orders.down.sql"}}},
		{"shared ancestry is not a collision", base, laneFiles{
			"a": with(base, map[string]string{d + "000002_a.up.sql": ""}),
			"b": with(base, map[string]string{d + "000002_a.up.sql": "", d + "000003_b.up.sql": ""}),
		}, want{}},
		{"lane collides with version merged to base", with(base, map[string]string{d + "000002_merged.up.sql": ""}), laneFiles{
			"a": with(base, map[string]string{d + "000002_late.up.sql": ""}),
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "000002_late.up.sql": d + "000003_late.up.sql"}}},
		{"lane older than base head", with(base, map[string]string{d + "000005_x.up.sql": ""}), laneFiles{
			"a": with(base, map[string]string{d + "000003_old.up.sql": ""}),
		}, want{kinds: []string{KindOutOfOrder}, renames: map[string]string{d + "000003_old.up.sql": d + "000006_old.up.sql"}}},
		{"unpadded duplicate", map[string]string{d + "1_init.up.sql": ""}, laneFiles{
			"a": {d + "1_init.up.sql": "", d + "2_a.up.sql": ""}, "b": {d + "1_init.up.sql": "", d + "02_b.up.sql": ""},
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "02_b.up.sql": d + "03_b.up.sql"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewGolangMigrate(), tc.base, tc.lanes), tc.want)
		})
	}
}
