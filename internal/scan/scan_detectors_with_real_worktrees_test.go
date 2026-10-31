package scan

import (
	"testing"

	"github.com/LysCongThanh/migrate-lanes/internal/testrepo"
)

// Each case: base committed on main, lane "agent-a" as an uncommitted
// worktree (agent still working), lane "agent-b" as a branch with no
// worktree. Both add a migration for the same next slot.
func TestScanEachDetectorAcrossRealWorktrees(t *testing.T) {
	gooseUp := "-- +goose Up\n"
	dbmateUp := "-- migrate:up\n"
	cases := []struct {
		tool, dir string
		base      map[string]string
		a, b      map[string]string
		kind      string
	}{
		{"golang-migrate", "db/migrations",
			map[string]string{"db/migrations/0001_init.up.sql": "", "db/migrations/0001_init.down.sql": ""},
			map[string]string{"db/migrations/0002_users.up.sql": "", "db/migrations/0002_users.down.sql": ""},
			map[string]string{"db/migrations/0002_orders.up.sql": "", "db/migrations/0002_orders.down.sql": ""}, "duplicate-version"},
		{"goose", "migrations",
			map[string]string{"migrations/00001_init.sql": gooseUp},
			map[string]string{"migrations/00002_a.sql": gooseUp},
			map[string]string{"migrations/00002_b.sql": gooseUp}, "duplicate-version"},
		{"sqlx", "migrations",
			map[string]string{"migrations/0001_init.sql": "create table t();"},
			map[string]string{"migrations/0002_a.sql": ""},
			map[string]string{"migrations/0002_b.sql": ""}, "duplicate-version"},
		{"dbmate", "db/migrations",
			map[string]string{"db/migrations/20240301000000_init.sql": dbmateUp},
			map[string]string{"db/migrations/20240401000000_a.sql": dbmateUp},
			map[string]string{"db/migrations/20240201000000_b.sql": dbmateUp}, "out-of-order"},
		{"rails", "db/migrate",
			map[string]string{"db/migrate/20240101000000_init.rb": ""},
			map[string]string{"db/migrate/20240102000000_a.rb": ""},
			map[string]string{"db/migrate/20240102000000_b.rb": ""}, "duplicate-version"},
		{"laravel", "database/migrations",
			map[string]string{"database/migrations/2024_01_01_000000_init.php": ""},
			map[string]string{"database/migrations/2024_01_02_000000_a.php": ""},
			map[string]string{"database/migrations/2024_01_02_000000_b.php": ""}, "duplicate-version"},
		{"prisma", "prisma/migrations",
			map[string]string{"prisma/migrations/20240101000000_init/migration.sql": ""},
			map[string]string{"prisma/migrations/20240102000000_a/migration.sql": ""},
			map[string]string{"prisma/migrations/20240102000000_b/migration.sql": ""}, "duplicate-version"},
		{"drizzle", "drizzle",
			map[string]string{"drizzle/meta/_journal.json": `{"entries":[{"idx":0,"tag":"0000_init"}]}`},
			map[string]string{"drizzle/meta/_journal.json": `{"entries":[{"idx":0,"tag":"0000_init"},{"idx":1,"tag":"0001_red"}]}`},
			map[string]string{"drizzle/meta/_journal.json": `{"entries":[{"idx":0,"tag":"0000_init"},{"idx":1,"tag":"0001_blue"}]}`}, "duplicate-version"},
		{"alembic", "alembic/versions",
			map[string]string{"alembic/versions/aaa_init.py": "revision = 'aaa'\ndown_revision = None\n"},
			map[string]string{"alembic/versions/bbb_a.py": "revision = 'bbb'\ndown_revision = 'aaa'\n"},
			map[string]string{"alembic/versions/ccc_b.py": "revision = 'ccc'\ndown_revision = 'aaa'\n"}, "multiple-heads"},
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			r := testrepo.New(t, tc.base)
			r.Worktree("agent-a", tc.a, false)
			r.Branch("agent-b", tc.b)
			res, err := Run(Options{Dir: r.Dir})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Targets) != 1 || res.Targets[0].Tool != tc.tool || res.Targets[0].Dir != tc.dir {
				t.Fatalf("targets = %+v, want %s at %s", res.Targets, tc.tool, tc.dir)
			}
			if len(res.Findings) != 1 || res.Findings[0].Kind != tc.kind {
				t.Fatalf("findings = %+v, want one %s", res.Findings, tc.kind)
			}
			lanes := res.Findings[0].Lanes()
			if tc.kind != "out-of-order" && (!has(lanes, "wt:agent-a") || !has(lanes, "agent-b")) {
				t.Fatalf("finding lanes = %v, want wt:agent-a and agent-b", lanes)
			}
		})
	}
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
