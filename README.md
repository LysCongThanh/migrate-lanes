# migrate-lanes

Catch database migration collisions **before merge** when several branches, worktrees or AI agents add migrations in parallel.

![migrate-lanes demo](docs/demo.gif)

```text
$ migrate-lanes scan
base main · 3 lane(s) · 1 migration dir(s)
  golang-migrate  db/migrations

1 collision(s):

[duplicate-version] golang-migrate db/migrations: version 0002 is used by 2 different migrations
    wt:agent-a                   db/migrations/0002_users.up.sql
    wt:agent-b                   db/migrations/0002_orders.up.sql
    fix (wt:agent-b): db/migrations/0002_orders.up.sql -> 0003_orders.up.sql
```

## The problem

Put three coding agents (or three developers) in three git worktrees and ask each one for a schema change. Each one looks at `db/migrations`, sees `0001`, and writes `0002`. Every branch passes CI by itself. The problem only shows up after merge:

- **Sequence tools** (golang-migrate, goose, sqlx) reject duplicate versions. A migration merged with a number lower than the database's current version never runs on databases that are already past it. [golang-migrate#470](https://github.com/golang-migrate/migrate/issues/470) asks for a way to catch exactly these concurrent additions.
- **drizzle-kit**: parallel branches generate migrations with the same id. The journal (`meta/_journal.json`) ends up with two entries at one `idx` and a broken snapshot chain, and today the fix is manual. See [drizzle-orm#2488](https://github.com/drizzle-team/drizzle-orm/issues/2488).
- **Alembic** ends up with two heads, and `alembic upgrade head` fails.
- **Timestamp tools** (dbmate, Rails, Laravel, Prisma) end up with duplicate timestamps, or with migrations merged out of order that run in a different order on fresh databases.

`migrate-lanes` looks at every **lane** (each git worktree as it is on disk, uncommitted files included, plus each local branch ahead of base) and reports these collisions before you merge. It can also suggest a renumbering and apply it.

## Install

```sh
go install github.com/LysCongThanh/migrate-lanes@latest
```

You can also download a prebuilt binary (with checksums) from the Releases page.

## Usage

```sh
migrate-lanes scan                      # all worktrees + local branches ahead of main
migrate-lanes scan --base develop --json
migrate-lanes scan --remotes            # also origin/* branches (CI: other open PRs)
migrate-lanes check db/migrations/0002_users.up.sql   # only collisions touching this file
migrate-lanes renumber                  # dry run: print suggested renames
migrate-lanes renumber --apply --lane wt:agent-b       # rename files inside that worktree
migrate-lanes hook claude               # print a Claude Code hook snippet
```

| Exit code | Meaning |
|---|---|
| 0 | clean |
| 1 | collisions found |
| 2 | usage / config / git error |

`check --hook` is the exception. It follows Claude Code hook semantics: exit 2 means "collision", and the message on stderr is fed back to the agent.

`renumber` never writes anything unless you pass `--apply`. It renames files only inside worktrees, it never overwrites a file, and it leaves committing to you. A branch without a worktree gets a suggestion only.

## Supported tools

| Tool | Detected by | Collisions reported |
|---|---|---|
| golang-migrate (and sqlx reversible) | `<n>_<name>.up.sql` / `.down.sql` | duplicate version, out-of-order |
| goose | `<n>_<name>.sql` with `-- +goose Up`, or `.go` calling `goose.AddMigration` | duplicate version, out-of-order |
| sqlx (simple) | `<n>_<name>.sql` | duplicate version, out-of-order |
| dbmate | `<YYYYMMDDHHMMSS>_<name>.sql` with `-- migrate:up` | duplicate timestamp, out-of-order |
| Rails | `db/migrate/<YYYYMMDDHHMMSS>_<name>.rb` | duplicate timestamp, out-of-order |
| Laravel | `YYYY_MM_DD_HHMMSS_<name>.php` | duplicate timestamp, out-of-order |
| Prisma | `migrations/<YYYYMMDDHHMMSS>_<name>/migration.sql` | duplicate timestamp, out-of-order |
| drizzle-kit | `<out>/meta/_journal.json` | same `idx` with different tags, journal diverged from base, unparsable journal |
| Alembic | `versions/*.py` with `revision` / `down_revision` | multiple heads, duplicate revision id |

Directories are detected automatically. Want another tool? It takes one detector file and one table test. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Config (optional)

`.migrate-lanes.toml` at the repo root:

```toml
base = "main"                       # default: main, then master
remotes = false                     # same as --remotes
ignore_branches = ["dependabot/*"]  # glob on branch names

[[paths]]                           # pin dirs instead of auto-detecting
dir = "db/migrations"
tool = "golang-migrate"             # optional; detected if omitted
```

## Claude Code hook

```sh
migrate-lanes hook claude
```

The command prints a `PostToolUse` hook for `Write|Edit|MultiEdit` that runs `migrate-lanes check --hook`. Merge it into `.claude/settings.json`. When an agent writes a migration that collides with another lane, the hook exits 2. Claude Code then shows the agent the collision and the suggested free version, and the agent can fix it right away instead of waiting for you to find it at merge time. For every other file the hook exits 0 quickly.

## GitHub Action

```yaml
on: pull_request
jobs:
  migrations:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: LysCongThanh/migrate-lanes@v0.1.0
        # with:
        #   base: main        # default: the PR's base branch
        #   remotes: "true"   # also compare with other branches on origin
```

The PR fails when its migrations collide with the base branch, or with another branch that has not been merged yet.

## Pairs with worktree-pg

`migrate-lanes` keeps parallel branches from colliding on **migration files**. [worktree-pg](https://github.com/LysCongThanh/worktree-pg) gives each worktree its **own Postgres database**, cloned from a migrated template. Use them together so parallel agents stop stepping on each other's schema.

## License

MIT © 2026 LysCongThanh
