package detect

import (
	"fmt"
	"testing"
)

// rev renders a minimal Alembic revision script (typed header style).
func rev(id, down string) string {
	d := "None"
	if down != "" {
		d = fmt.Sprintf("%q", down)
	}
	return fmt.Sprintf("revision: str = %q\ndown_revision: Union[str, None] = %s\n\ndef upgrade():\n    pass\n", id, d)
}

func TestAlembic(t *testing.T) {
	const d = "alembic/versions/"
	base := map[string]string{d + "aaa_init.py": rev("aaa", "")}
	cases := []struct {
		name  string
		lanes laneFiles
		want  want
	}{
		{"linear chain", laneFiles{"a": with(base, map[string]string{d + "bbb_a.py": rev("bbb", "aaa")})}, want{}},
		{"stacked lanes stay linear", laneFiles{
			"a": with(base, map[string]string{d + "bbb_a.py": rev("bbb", "aaa")}),
			"b": with(base, map[string]string{d + "bbb_a.py": rev("bbb", "aaa"), d + "ccc_b.py": rev("ccc", "bbb")}),
		}, want{}},
		{"two lanes branch from the same head", laneFiles{
			"a": with(base, map[string]string{d + "bbb_a.py": rev("bbb", "aaa")}),
			"b": with(base, map[string]string{d + "ccc_b.py": rev("ccc", "aaa")}),
		}, want{kinds: []string{KindHeads}}},
		{"merge revision with tuple down_revision", laneFiles{
			"a": with(base, map[string]string{d + "bbb_a.py": rev("bbb", "aaa"), d + "ccc_b.py": rev("ccc", "aaa"),
				d + "ddd_merge.py": "revision = 'ddd'\ndown_revision = ('bbb', 'ccc')\n"}),
		}, want{}},
		{"duplicate revision id", laneFiles{
			"a": with(base, map[string]string{d + "bbb_a.py": rev("bbb", "aaa")}),
			"b": with(base, map[string]string{d + "bbb_other.py": rev("bbb", "aaa")}),
		}, want{kinds: []string{KindDuplicate}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, runDetector(t, NewAlembic(), base, tc.lanes), tc.want)
		})
	}
}
