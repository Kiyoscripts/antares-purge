package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPatchPreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	original := `agent:
    max_turns: 200
    workspace: /home/me/work
gateway:
    telegram:
        token: secret-token
tools:
    approval_mode: prompt
`
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	var captured string
	err := Patch(path, PatchSpec{
		Toolset:      "security",
		ApprovalMode: "auto",
		SkillsDir:    "/home/me/.antares/skills",
		PluginsDir:   "/home/me/.antares/plugins",
		RolesDir:     "/home/me/.antares/roles",
	}, false, func(p string) error { captured = p; return nil })
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if captured != path {
		t.Error("capture hook not called with the config path")
	}

	raw, _ := os.ReadFile(path)
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("result is not valid YAML: %v", err)
	}

	// Unknown keys survive.
	agent := root["agent"].(map[string]any)
	if agent["max_turns"] != 200 {
		t.Errorf("agent.max_turns lost: %v", agent["max_turns"])
	}
	gw := root["gateway"].(map[string]any)
	tg := gw["telegram"].(map[string]any)
	if tg["token"] != "secret-token" {
		t.Errorf("gateway.telegram.token lost: %v", tg["token"])
	}

	// Our fields are set.
	tools := root["tools"].(map[string]any)
	if tools["toolset"] != "security" {
		t.Errorf("toolset = %v, want security", tools["toolset"])
	}
	if tools["approval_mode"] != "auto" {
		t.Errorf("approval_mode = %v, want auto", tools["approval_mode"])
	}
}

func TestPatchIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	spec := PatchSpec{Toolset: "security", ApprovalMode: "auto", SkillsDir: "/s", PluginsDir: "/p", RolesDir: "/r"}

	if err := Patch(path, spec, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := Patch(path, spec, false, nil); err != nil {
		t.Fatal(err)
	}

	raw, _ := os.ReadFile(path)
	var root map[string]any
	yaml.Unmarshal(raw, &root)
	dirs := root["skills"].(map[string]any)["dirs"].([]any)
	if len(dirs) != 1 {
		t.Errorf("skills.dirs duplicated on re-patch: %v", dirs)
	}
}

func TestPatchRejectsInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("this: : : not yaml\n\t- broken"), 0o600)
	err := Patch(path, PatchSpec{Toolset: "security"}, false, nil)
	if err == nil || !strings.Contains(err.Error(), "valid YAML") {
		t.Errorf("expected a YAML error, got %v", err)
	}
}
