package deep

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The patch is the reason a purge install gets a working turn_end hook at all:
// upstream declares plugin.TurnEnd and never dispatches it. This drives the
// patch against a pristine source tree and asserts both the result and that a
// second run is a no-op.
func TestPatchTurnEndOnRealTree(t *testing.T) {
	root := t.TempDir()
	arc, err := exec.Command("git", "-C", "/tmp/antares-src", "archive", "HEAD").Output()
	if err != nil {
		t.Skipf("no pristine source tree available: %v", err)
	}

	// Pipe the archive into tar's stdin rather than using a shell redirect.
	tarCmd := exec.Command("tar", "xf", "-", "-C", root)
	tarCmd.Stdin = strings.NewReader(string(arc))
	if out, err := tarCmd.CombinedOutput(); err != nil {
		t.Fatalf("extract: %v: %s", err, out)
	}

	if err := patchTurnEnd(root); err != nil {
		t.Fatalf("patchTurnEnd: %v", err)
	}
	if err := patchTurnEnd(root); err != nil {
		t.Fatalf("second run must be idempotent: %v", err)
	}

	harness, _ := os.ReadFile(filepath.Join(root, "internal", "agent", "harness.go"))
	if !strings.Contains(string(harness), "notifyPluginsTurnEnd") {
		t.Fatal("harness.go: helper not added")
	}
	life, _ := os.ReadFile(filepath.Join(root, "internal", "agent", "turn_lifecycle.go"))
	if !strings.Contains(string(life), "notifyPluginsTurnEnd") {
		t.Fatal("turn_lifecycle.go: dispatch not added")
	}
	if !strings.Contains(string(life), "/internal/plugin\"") {
		t.Fatal("turn_lifecycle.go: plugin import not added")
	}
	agent, _ := os.ReadFile(filepath.Join(root, "internal", "agent", "agent.go"))
	if !strings.Contains(string(agent), "lastReply, turn, failures") {
		t.Fatal("agent.go: turn counter not threaded")
	}
	fmt.Printf("OK: patched and idempotent over a %d-byte tree\n", len(arc))
}
