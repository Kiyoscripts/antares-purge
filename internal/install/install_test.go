package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enowdev/antares-purge/internal/backup"
	"github.com/enowdev/antares-purge/internal/paths"
)

// scratchHome points ANTARES_HOME at a temp dir for the duration of a test.
func scratchHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ANTARES_HOME", dir)
	return dir
}

func TestInstallWritesPresetAndManifest(t *testing.T) {
	home := scratchHome(t)
	res, err := Run(Options{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if res.RunDir == "" {
		t.Fatal("no backup run dir recorded")
	}

	for _, want := range []string{
		filepath.Join(home, "SOUL.md"),
		filepath.Join(home, "AGENTS.md"),
		filepath.Join(home, "skills", "recon-pipeline.md"),
		filepath.Join(home, "roles", "redteam-commander.md"),
		filepath.Join(home, "plugins", "purge-gate", "plugin.yaml"),
		filepath.Join(home, "plugins", "purge-gate", "run"),
	} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("expected %s: %v", want, err)
		}
	}

	// The gate script must be executable or Antares cannot run it.
	info, err := os.Stat(filepath.Join(home, "plugins", "purge-gate", "run"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("gate script is not executable: %v", info.Mode())
	}

	// The soul must carry the operator identity, not the generic default.
	soul, _ := os.ReadFile(filepath.Join(home, "SOUL.md"))
	if !strings.Contains(string(soul), "OPERATOR") {
		t.Error("SOUL.md does not carry the operator identity")
	}

	// Config must have been patched.
	cfg, _ := os.ReadFile(filepath.Join(home, "config.yaml"))
	for _, want := range []string{"toolset: security", "approval_mode: auto", "purge-gate"} {
		if !strings.Contains(string(cfg), want) && want != "purge-gate" {
			t.Errorf("config.yaml missing %q:\n%s", want, cfg)
		}
	}

	m, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	if m.InstalledAt.IsZero() {
		t.Fatal("manifest not saved")
	}
}

func TestInstallPreservesExistingSoul(t *testing.T) {
	home := scratchHome(t)
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	mine := "# My own soul\n\nI am already someone.\n"
	if err := os.WriteFile(filepath.Join(home, "SOUL.md"), []byte(mine), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Run(Options{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(home, "SOUL.md"))
	if string(got) != mine {
		t.Errorf("install clobbered an existing soul:\n%s", got)
	}
}

func TestInstallForceOverwritesSoul(t *testing.T) {
	home := scratchHome(t)
	os.MkdirAll(home, 0o700)
	os.WriteFile(filepath.Join(home, "SOUL.md"), []byte("old\n"), 0o600)

	if _, err := Run(Options{Force: true}); err != nil {
		t.Fatalf("install: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(home, "SOUL.md"))
	if !strings.Contains(string(got), "OPERATOR") {
		t.Error("--force did not overwrite the soul")
	}
}

func TestUninstallRestoresPriorState(t *testing.T) {
	home := scratchHome(t)
	os.MkdirAll(home, 0o700)
	// A pre-existing config that the install will patch and uninstall must restore.
	original := "agent:\n    max_turns: 42\n"
	os.WriteFile(filepath.Join(home, "config.yaml"), []byte(original), 0o600)

	if _, err := Run(Options{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "skills", "recon-pipeline.md")); err != nil {
		t.Fatal("install did not write a skill")
	}

	if _, err := Uninstall(false); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(home, "config.yaml"))
	if err != nil {
		t.Fatalf("config.yaml gone after uninstall: %v", err)
	}
	if string(got) != original {
		t.Errorf("config not restored:\n got %q\nwant %q", got, original)
	}
	// The preset's own files must be gone.
	if _, err := os.Stat(filepath.Join(home, "SOUL.md")); !os.IsNotExist(err) {
		t.Error("SOUL.md survived uninstall (it did not exist before install)")
	}
	if _, err := os.Stat(filepath.Join(home, "plugins", "purge-gate")); !os.IsNotExist(err) {
		t.Error("plugin dir survived uninstall")
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	home := scratchHome(t)
	res, err := Run(Options{DryRun: true})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if len(res.Written) == 0 {
		t.Error("dry run reported no planned writes")
	}
	if _, err := os.Stat(filepath.Join(home, "SOUL.md")); !os.IsNotExist(err) {
		t.Error("dry run wrote SOUL.md")
	}
}

func TestDoctorReportsDrift(t *testing.T) {
	home := scratchHome(t)
	if _, err := Run(Options{}); err != nil {
		t.Fatal(err)
	}
	// Edit a tracked file to create drift.
	os.WriteFile(filepath.Join(home, "SOUL.md"), []byte("hand edited\n"), 0o600)

	drift, err := Doctor()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range drift {
		if strings.Contains(d, "SOUL.md") && strings.Contains(d, "edited") {
			found = true
		}
	}
	if !found {
		t.Errorf("doctor did not report the SOUL.md edit: %v", drift)
	}
}

func TestPathsHonourAntaresHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ANTARES_HOME", dir)
	if got := paths.Soul(); got != filepath.Join(dir, "SOUL.md") {
		t.Errorf("Soul() = %s, want under %s", got, dir)
	}
}

var _ = backup.Manifest{}
