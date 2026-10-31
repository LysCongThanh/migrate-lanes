package detect

import "testing"

func TestLaravel(t *testing.T) {
	const d = "database/migrations/"
	base := map[string]string{d + "2024_01_01_000000_create_users_table.php": "", d + "2024_03_01_000000_x.php": ""}
	cases := []struct {
		name  string
		lanes laneFiles
		want  want
	}{
		{"clean", laneFiles{"a": with(base, map[string]string{d + "2024_04_01_000000_a.php": ""})}, want{}},
		{"out of order keeps laravel format", laneFiles{"a": with(base, map[string]string{d + "2024_02_01_000000_a.php": ""})},
			want{kinds: []string{KindOutOfOrder}, renames: map[string]string{d + "2024_02_01_000000_a.php": d + "2024_03_01_000001_a.php"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewLaravel(), base, tc.lanes), tc.want)
		})
	}
}
