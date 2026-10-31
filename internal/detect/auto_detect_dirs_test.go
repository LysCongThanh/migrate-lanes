package detect

import (
	"testing"

	"github.com/LysCongThanh/migrate-lanes/internal/lane"
)

func TestAutoDetectPicksMostSpecificTool(t *testing.T) {
	l := lane.FromMemory("base", map[string]string{
		"a/migrations/0001_x.up.sql":                         "",
		"b/migrations/20240101000000_x.sql":                  "-- migrate:up\n",
		"c/migrations/00001_x.sql":                           "-- +goose Up\n",
		"d/migrations/0001_x.sql":                            "create table t();",
		"e/drizzle/meta/_journal.json":                       `{"entries":[]}`,
		"e/drizzle/0000_x.sql":                               "",
		"f/alembic/versions/aaa.py":                          "revision = 'aaa'\ndown_revision = None\n",
		"g/prisma/migrations/20240101000000_x/migration.sql": "",
		"h/database/migrations/2024_01_01_000000_x.php":      "",
		"i/db/migrate/20240101000000_x.rb":                   "",
		"docs/1_intro.md":                                    "",
	})
	want := map[string]string{
		"a/migrations": "golang-migrate", "b/migrations": "dbmate", "c/migrations": "goose", "d/migrations": "sqlx",
		"e/drizzle": "drizzle", "f/alembic/versions": "alembic", "g/prisma/migrations": "prisma",
		"h/database/migrations": "laravel", "i/db/migrate": "rails",
	}
	got := AutoDetect([]*lane.Lane{l})
	if len(got) != len(want) {
		t.Fatalf("got %d targets %+v, want %d", len(got), got, len(want))
	}
	for _, tg := range got {
		if want[tg.Dir] != tg.Tool {
			t.Errorf("%s: tool %s, want %s", tg.Dir, tg.Tool, want[tg.Dir])
		}
	}
}
