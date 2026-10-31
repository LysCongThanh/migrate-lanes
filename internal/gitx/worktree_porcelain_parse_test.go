package gitx

import "testing"

func TestParseWorktreePorcelain(t *testing.T) {
	out := "worktree /repo\nHEAD abc\nbranch refs/heads/main\n\n" +
		"worktree /wt/feat\nHEAD def\nbranch refs/heads/feat/x\nlocked reason\n\n" +
		"worktree /wt/det\nHEAD 123\ndetached\nprunable gitdir file points to non-existent location\n\n" +
		"worktree /bare.git\nbare\n"
	got := parseWorktreePorcelain(out)
	if len(got) != 4 {
		t.Fatalf("got %d worktrees", len(got))
	}
	if got[1].Branch != "feat/x" || got[1].Path != "/wt/feat" {
		t.Fatalf("feat = %+v", got[1])
	}
	if got[2].Branch != "" || !got[2].Prunable || !got[3].Bare {
		t.Fatalf("detached/bare = %+v %+v", got[2], got[3])
	}
}
