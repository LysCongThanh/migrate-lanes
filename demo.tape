# vhs demo.tape  ->  demo.gif  (https://github.com/charmbracelet/vhs)
Output demo.gif
Set FontSize 16
Set Width 1100
Set Height 620
Set Theme "Catppuccin Mocha"

Hide
Type "cd $(mktemp -d) && git init -q -b main app && cd app && mkdir -p db/migrations && touch db/migrations/0001_init.up.sql && git add -A && git commit -qm init" Enter
Type "git worktree add -q -b agent-a ../agent-a && git worktree add -q -b agent-b ../agent-b && clear" Enter
Show

Type "# two agents, two worktrees, both add migration 0002" Enter
Type "touch ../agent-a/db/migrations/0002_users.up.sql ../agent-b/db/migrations/0002_orders.up.sql" Enter
Sleep 500ms
Type "migrate-lanes scan" Enter
Sleep 3s
Type "migrate-lanes renumber --apply --lane wt:agent-b" Enter
Sleep 2s
Type "migrate-lanes scan" Enter
Sleep 3s
