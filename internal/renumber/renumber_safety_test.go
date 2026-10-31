package renumber

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LysCongThanh/migrate-lanes/internal/detect"
	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

func TestApplyRefusesOverwriteAndEscape(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{"m/0002_a.sql", "m/0003_a.sql"} {
		p := filepath.Join(root, f)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	wt := &lane.Lane{Name: "wt:x", Path: root}
	lookup := func(string) *lane.Lane { return wt }
	overwrite := []Step{{Rename: detect.Rename{Lane: "wt:x", From: "m/0002_a.sql", To: "m/0003_a.sql"}, Applicable: true}}
	if _, err := Apply(overwrite, lookup); err == nil {
		t.Fatal("expected overwrite refusal")
	}
	escape := []Step{{Rename: detect.Rename{Lane: "wt:x", From: "m/0002_a.sql", To: "../../evil.sql"}, Applicable: true}}
	if _, err := Apply(escape, lookup); err == nil {
		t.Fatal("expected escape refusal")
	}
}

func TestPlanMarksBranchOnlyLanesNotApplicable(t *testing.T) {
	f := detect.Finding{Renames: []detect.Rename{{Lane: "feat", From: "m/1_a.sql", To: "m/2_a.sql"}}}
	branch := &lane.Lane{Name: "feat", Kind: lane.KindBranch}
	steps, advice := Plan([]detect.Finding{f, {Kind: detect.KindHeads, Entries: []detect.Entry{{Lane: "feat"}}}},
		func(string) *lane.Lane { return branch }, "")
	if len(steps) != 1 || steps[0].Applicable || len(advice) != 1 {
		t.Fatalf("steps %+v advice %+v", steps, advice)
	}
}
