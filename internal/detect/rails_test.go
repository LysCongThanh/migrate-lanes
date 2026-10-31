package detect

import "testing"

func TestRails(t *testing.T) {
	const d = "db/migrate/"
	base := map[string]string{d + "20240101000000_create_users.rb": ""}
	cases := []struct {
		name  string
		lanes laneFiles
		want  want
	}{
		{"clean", laneFiles{"a": with(base, map[string]string{d + "20240102000000_add_email.rb": ""})}, want{}},
		{"duplicate timestamp", laneFiles{
			"a": with(base, map[string]string{d + "20240102000000_add_email.rb": ""}),
			"b": with(base, map[string]string{d + "20240102000000_add_phone.rb": ""}),
		}, want{kinds: []string{KindDuplicate}, renames: map[string]string{d + "20240102000000_add_phone.rb": d + "20240102000001_add_phone.rb"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewRails(), base, tc.lanes), tc.want)
		})
	}
}
