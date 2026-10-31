package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LysCongThanh/migrate-lanes/internal/testrepo"
)

// run executes the CLI in-process and captures streams.
func run(stdin string, args ...string) (code int, stdout, stderr string) {
	var out, errb bytes.Buffer
	code = Main(args, IO{In: strings.NewReader(stdin), Out: &out, Err: &errb})
	return code, out.String(), errb.String()
}

// collidingRepo: worktrees a and b both add version 0002.
func collidingRepo(t *testing.T) (r *testrepo.Repo, wtA, wtB string) {
	r = testrepo.New(t, map[string]string{"db/migrations/0001_init.up.sql": "", "db/migrations/0001_init.down.sql": ""})
	wtA = r.Worktree("a", map[string]string{"db/migrations/0002_users.up.sql": "", "db/migrations/0002_users.down.sql": ""}, true)
	wtB = r.Worktree("b", map[string]string{"db/migrations/0002_orders.up.sql": "", "db/migrations/0002_orders.down.sql": ""}, false)
	return r, wtA, wtB
}

func TestScanExitCodesAndJSON(t *testing.T) {
	r := testrepo.New(t, map[string]string{"db/migrations/0001_init.up.sql": ""})
	if code, out, _ := run("", "scan", "-C", r.Dir); code != ExitClean || !strings.Contains(out, "no collisions") {
		t.Fatalf("clean scan: code %d out %q", code, out)
	}
	r2, _, _ := collidingRepo(t)
	code, out, _ := run("", "scan", "--json", "-C", r2.Dir)
	if code != ExitCollision {
		t.Fatalf("collision scan code = %d", code)
	}
	var res struct {
		Findings []struct{ Kind string } `json:"findings"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil || len(res.Findings) != 1 {
		t.Fatalf("json: %v %q", err, out)
	}
	if code, _, _ := run("", "scan", "--base", "missing", "-C", r2.Dir); code != ExitUsage {
		t.Fatalf("bad base code = %d", code)
	}
	if code, _, _ := run("", "bogus"); code != ExitUsage {
		t.Fatalf("unknown command code = %d", code)
	}
}

func TestCheckHookWarnsAgentOnStderr(t *testing.T) {
	_, _, wtB := collidingRepo(t)
	file := filepath.Join(wtB, "db/migrations/0002_orders.up.sql")
	payload := `{"hook_event_name":"PostToolUse","cwd":"` + wtB + `","tool_name":"Write","tool_input":{"file_path":"` + file + `"}}`
	code, out, errOut := run(payload, "check", "--hook")
	if code != 2 || out != "" || !strings.Contains(errOut, "0003_orders.up.sql") {
		t.Fatalf("hook: code %d stdout %q stderr %q", code, out, errOut)
	}
	// Non-migration file and unrelated payloads stay silent.
	readme := `{"cwd":"` + wtB + `","tool_input":{"file_path":"README.md"}}`
	if code, _, e := run(readme, "check", "--hook"); code != 0 || e != "" {
		t.Fatalf("readme: code %d stderr %q", code, e)
	}
	if code, _, _ := run(`{"tool_input":{"command":"ls"}}`, "check", "--hook"); code != 0 {
		t.Fatalf("bash payload code = %d", code)
	}
	outside := `{"tool_input":{"file_path":"` + filepath.Join(t.TempDir(), "x.sql") + `"}}`
	if code, _, _ := run(outside, "check", "--hook"); code != 0 {
		t.Fatalf("outside repo code = %d", code)
	}
}

func TestCheckFileModeExitCodes(t *testing.T) {
	_, _, wtB := collidingRepo(t)
	if code, _, _ := run("", "check", filepath.Join(wtB, "db/migrations/0002_orders.up.sql")); code != ExitCollision {
		t.Fatalf("check colliding file code = %d", code)
	}
	if code, _, _ := run("", "check", filepath.Join(wtB, "README.md")); code != ExitClean {
		t.Fatalf("check unrelated file code = %d", code)
	}
}

func TestRenumberDryRunThenApply(t *testing.T) {
	_, _, wtB := collidingRepo(t)
	code, out, _ := run("", "renumber", "-C", wtB)
	if code != ExitCollision || !strings.Contains(out, "dry run") || !strings.Contains(out, "0003_orders.up.sql") {
		t.Fatalf("dry run: code %d out %q", code, out)
	}
	if _, err := os.Stat(filepath.Join(wtB, "db/migrations/0002_orders.up.sql")); err != nil {
		t.Fatal("dry run must not touch files")
	}
	if code, out, errOut := run("", "renumber", "--apply", "--lane", "wt:b", "-C", wtB); code != ExitClean {
		t.Fatalf("apply: code %d out %q err %q", code, out, errOut)
	}
	for _, f := range []string{"0003_orders.up.sql", "0003_orders.down.sql"} {
		if _, err := os.Stat(filepath.Join(wtB, "db/migrations", f)); err != nil {
			t.Fatalf("expected %s after apply: %v", f, err)
		}
	}
	if code, _, _ := run("", "scan", "-C", wtB); code != ExitClean {
		t.Fatalf("scan after apply code = %d", code)
	}
	if code, _, _ := run("", "renumber", "--apply", "--dry-run"); code != ExitUsage {
		t.Fatalf("conflicting flags code = %d", code)
	}
}

func TestHookClaudeSnippetIsValidJSON(t *testing.T) {
	code, out, _ := run("", "hook", "claude")
	var v map[string]any
	if code != 0 || json.Unmarshal([]byte(out), &v) != nil || !strings.Contains(out, "check --hook") {
		t.Fatalf("hook claude: code %d out %q", code, out)
	}
	if code, _, _ := run("", "hook", "cursor"); code != ExitUsage {
		t.Fatalf("unknown hook target code = %d", code)
	}
}
