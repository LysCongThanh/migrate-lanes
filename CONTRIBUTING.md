# Contributing

Thanks for helping. The most useful contribution is a **detector** for another migration tool. A detector is one file plus one table test, so it makes a good first PR.

## Dev setup

```sh
go test ./...        # needs git on PATH; tests create temp repos with real worktrees
go vet ./...
gofmt -l .           # must print nothing
```

House rules:
- Go files stay under 200 lines and use snake_case file names.
- Add a comment that explains *why*.
- No mocks of git: tests build real repos (`internal/testrepo`).

## Add a detector (good first issue)

1. **Pick the shape.** If the tool names migrations `<version>_<name>...`, which covers most of them, you only fill in a spec. Copy `internal/detect/rails.go`:

   ```go
   // Knex: migrations/20240101120000_create_users.js
   func NewKnex() Detector {
       return &versionedDetector{
           tool:    "knex",
           pattern: regexp.MustCompile(`^(\d{14})(_.+)\.(?:js|ts|cjs|mjs)$`), // group 1 = version prefix (must start the name), group 2 = rest of the identity
           advice:  "give the lane's migration a newer timestamp",
       }
   }
   ```

   Optional fields:
   - `marker`: a content regex that must appear before a directory is claimed. Example: dbmate's `-- migrate:up`.
   - `folderFile`: set it when each migration is a folder (Prisma).
   - `formatPrefix`: set it when the version has separators (Laravel).

   You get duplicate and out-of-order detection, plus rename suggestions, with no extra code.

2. **Graph or journal tools** (like `drizzle.go` and `alembic.go`) implement the `Detector` interface directly:
   - `Tool()`
   - `Claim(lane)` returns the directories the tool owns.
   - `Analyze(dir, base, lanes)` returns `[]Finding`.

3. **Register** the detector in `All()` in `internal/detect/detector.go`. Order matters: the more specific layout goes first.

4. **Table test** `internal/detect/<tool>_test.go`: copy `rails_test.go` and cover at least the clean case, a duplicate, and an out-of-order case. Use `runDetector` and `with` from `helpers_for_detector_tests_test.go`.

5. **Real-worktree case:** add one row to `internal/scan/scan_detectors_with_real_worktrees_test.go`.

6. Add a row to the Supported tools table in `README.md`.

Ideas: Knex, Flyway (needs a prefix-aware rename for `V<n>__`), Liquibase changelogs, Atlas (`atlas.sum`), TypeORM, Sequelize, Django (`dependencies` graph), Ecto, Diesel, EF Core.

## Commits and PRs

- Use conventional commits (`feat:`, `fix:`, `test:`, `docs:`).
- Keep each PR to one topic.
- CI runs vet, gofmt, `go test -race` on Linux and macOS, and gitleaks.
