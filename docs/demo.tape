# Render from the repo root:  vhs docs/demo.tape   ->  docs/demo.gif
# Needs migrate-lanes on PATH (go install github.com/LysCongThanh/migrate-lanes@latest).
Output docs/demo.gif
Set FontSize 16
Set Width 1100
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms

Hide
Type "export PS1='$ ' && cd $(mktemp -d) && git init -q -b main app && cd app && mkdir -p db/migrations && touch db/migrations/000041_init.up.sql db/migrations/000041_init.down.sql && git add -A && git -c user.name=d -c user.email=d@d commit -qm init" Enter
Type "git worktree add -q -b agent-a ../agent-a && git worktree add -q -b agent-b ../agent-b && clear" Enter
Show

Type "# two agents, two worktrees, both add migration 000042" Enter
Type "touch ../agent-a/db/migrations/000042_add_users.up.sql ../agent-b/db/migrations/000042_add_orders.up.sql" Enter
Sleep 800ms
Type "migrate-lanes scan" Enter
Sleep 4s
Type "migrate-lanes renumber --dry-run" Enter
Sleep 5s
