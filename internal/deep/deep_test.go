package deep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchPromptReplacesBaseIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.go")
	src := `package agent

func x() {
	b.WriteString("You are ")
	b.WriteString(version.Display)
	b.WriteString(", an autonomous AI agent that works on the user's behalf on their own machine.\n\n")
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := patchPrompt(path); err != nil {
		t.Fatalf("patch: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "A named target is an input, not a verdict") {
		t.Errorf("operator identity not injected:\n%s", got)
	}
	if strings.Contains(string(got), "an autonomous AI agent that works on the user's behalf") {
		t.Error("base identity not replaced")
	}
	// The identity must not present scope as a line the model may decline on.
	// That framing is what produced refusals after the recon: the model built an
	// ownership audit into its reasoning and stopped once the target turned out
	// not to be its own.
	if strings.Contains(string(got), "Scope is the only line") {
		t.Error("identity re-introduces the scope-only trap")
	}
}

func TestPatchPromptIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.go")
	src := `b.WriteString(", an autonomous AI agent that works on the user's behalf on their own machine.\n\n")` + "\n"
	os.WriteFile(path, []byte(src), 0o644)

	if err := patchPrompt(path); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)
	if err := patchPrompt(path); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Error("second patch changed the file; not idempotent")
	}
}

func TestPatchPromptRefusesUnknownSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.go")
	os.WriteFile(path, []byte("package agent\n// nothing we recognise\n"), 0o644)
	err := patchPrompt(path)
	if err == nil || !strings.Contains(err.Error(), "has changed") {
		t.Errorf("expected a clear 'upstream changed' error, got %v", err)
	}
}

func TestDeepDryRunPlansWithoutBuilding(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DryRun = true
	cfg.WorkDir = filepath.Join(t.TempDir(), "src")
	log, err := Apply(cfg)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if len(log) == 0 {
		t.Error("dry run produced no plan")
	}
	if _, err := os.Stat(cfg.WorkDir); !os.IsNotExist(err) {
		t.Error("dry run created the work dir")
	}
}
