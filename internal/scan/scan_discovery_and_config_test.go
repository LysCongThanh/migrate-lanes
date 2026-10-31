package scan

import (
	"errors"
	"strings"
	"testing"

	"github.com/LysCongThanh/migrate-lanes/internal/testrepo"
)

var mBase = map[string]string{"db/migrations/0001_init.up.sql": ""}

func laneNames(res *Result) []string {
	var out []string
	for _, l := range res.Lanes {
		out = append(out, l.Name)
	}
	return out
}

func TestDiscoverySkipsMergedAndIgnoredBranches(t *testing.T) {
	r := testrepo.New(t, mBase)
	r.Worktree("feat/a", map[string]string{"db/migrations/0002_a.up.sql": ""}, true)
	r.Branch("feat/b", map[string]string{"db/migrations/0003_b.up.sql": ""})
	r.Git(r.Dir, "branch", "merged-already", "main") // not ahead of main
	r.Branch("dependabot/x", map[string]string{"go.sum": "x"})
	r.Write(r.Dir, map[string]string{".migrate-lanes.toml": "ignore_branches = [\"dependabot/*\"]\n"})
	res, err := Run(Options{Dir: r.Dir})
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(laneNames(res), ",")
	if got != "wt:main,wt:feat/a,feat/b" {
		t.Fatalf("lanes = %s", got)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("unexpected findings %+v", res.Findings)
	}
}

func TestUncommittedFileInMainWorktreeCollides(t *testing.T) {
	r := testrepo.New(t, mBase)
	r.Worktree("agent", map[string]string{"db/migrations/0002_agent.up.sql": ""}, true)
	r.Write(r.Dir, map[string]string{"db/migrations/0002_local.up.sql": ""}) // not committed
	res, err := Run(Options{Dir: r.Dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 || len(res.Findings[0].Renames) != 1 {
		t.Fatalf("findings = %+v", res.Findings)
	}
}

func TestConfigPinsPathAndTool(t *testing.T) {
	r := testrepo.New(t, map[string]string{
		"schema/0001_init.sql": "",
		".migrate-lanes.toml":  "base = \"main\"\n[[paths]]\ndir = \"schema/\"\ntool = \"sqlx\"\n",
	})
	r.Worktree("a", map[string]string{"schema/0002_a.sql": ""}, false)
	r.Worktree("b", map[string]string{"schema/0002_b.sql": ""}, false)
	res, err := Run(Options{Dir: r.Dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Targets) != 1 || res.Targets[0].Dir != "schema" || len(res.Findings) != 1 {
		t.Fatalf("targets %+v findings %+v", res.Targets, res.Findings)
	}
}

func TestSetupErrors(t *testing.T) {
	r := testrepo.New(t, mBase)
	if _, err := Run(Options{Dir: r.Dir, Base: "nope"}); err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("missing base: err = %v", err)
	}
	if _, err := Run(Options{Dir: t.TempDir()}); !errors.Is(err, ErrNotRepo) {
		t.Fatalf("outside repo: err = %v", err)
	}
	r.Write(r.Dir, map[string]string{".migrate-lanes.toml": "[[paths]]\ndir = \"../x\"\n"})
	if _, err := Run(Options{Dir: r.Dir}); err == nil || !strings.Contains(err.Error(), "inside the repo") {
		t.Fatalf("escaping path: err = %v", err)
	}
	r.Write(r.Dir, map[string]string{".migrate-lanes.toml": "bse = \"main\"\n"})
	if _, err := Run(Options{Dir: r.Dir}); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("typo key: err = %v", err)
	}
}

func TestRemoteBranchesAreLanesWithRemotesFlag(t *testing.T) {
	r := testrepo.New(t, mBase)
	r.Branch("pr-1", map[string]string{"db/migrations/0002_pr1.up.sql": ""})
	r.Branch("pr-2", map[string]string{"db/migrations/0002_pr2.up.sql": ""})
	// Simulate CI: remote-tracking refs exist, local branches do not.
	for _, b := range []string{"main", "pr-1", "pr-2"} {
		r.Git(r.Dir, "update-ref", "refs/remotes/origin/"+b, b)
	}
	r.Git(r.Dir, "branch", "-D", "pr-1", "pr-2")
	res, err := Run(Options{Dir: r.Dir, Base: "origin/main", Remotes: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(laneNames(res), ","); got != "wt:main,origin/pr-1,origin/pr-2" {
		t.Fatalf("lanes = %s", got)
	}
	if len(res.Findings) != 1 {
		t.Fatalf("findings = %+v", res.Findings)
	}
}
