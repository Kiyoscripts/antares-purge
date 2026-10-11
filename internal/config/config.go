// Package config patches ~/.antares/config.yaml for the preset without
// clobbering anything it does not own. It round-trips the file through a generic
// YAML tree so comments-as-keys, profiles and future fields survive a write.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PatchSpec is the set of fields the preset wants set.
type PatchSpec struct {
	Toolset      string
	ApprovalMode string
	SkillsDir    string
	PluginsDir   string
	RolesDir     string
}

// Patch reads path, applies spec, and writes it back. capture is called with
// the path before any write, so the caller can back it up.
func Patch(path string, spec PatchSpec, dryRun bool, capture func(string) error) error {
	root := map[string]any{}
	raw, err := os.ReadFile(path)
	switch {
	case err == nil:
		if len(strings.TrimSpace(string(raw))) > 0 {
			if err := yaml.Unmarshal(raw, &root); err != nil {
				return fmt.Errorf("config.yaml is not valid YAML: %w", err)
			}
		}
	case os.IsNotExist(err):
		// A fresh machine: build a minimal tree.
	default:
		return err
	}

	setNested(root, []string{"tools", "toolset"}, spec.Toolset)
	setNested(root, []string{"tools", "approval_mode"}, spec.ApprovalMode)
	appendUnique(root, []string{"skills", "dirs"}, spec.SkillsDir)
	appendUnique(root, []string{"plugins", "dirs"}, spec.PluginsDir)
	appendUnique(root, []string{"roles", "dirs"}, spec.RolesDir)
	setNested(root, []string{"skills", "enabled"}, true)
	setNested(root, []string{"plugins", "enabled"}, true)

	if dryRun {
		return nil
	}
	if capture != nil {
		if err := capture(path); err != nil {
			return err
		}
	}

	out, err := yaml.Marshal(root)
	if err != nil {
		return err
	}
	tmp := path + ".purge.tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// setNested sets root[a][b]... = value, creating intermediate maps.
func setNested(root map[string]any, keys []string, value any) {
	cur := root
	for _, k := range keys[:len(keys)-1] {
		next, ok := cur[k].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[k] = next
		}
		cur = next
	}
	cur[keys[len(keys)-1]] = value
}

// appendUnique adds value to the list at root[a][b] if not already present.
func appendUnique(root map[string]any, keys []string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	cur := root
	for _, k := range keys[:len(keys)-1] {
		next, ok := cur[k].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[k] = next
		}
		cur = next
	}
	last := keys[len(keys)-1]
	var list []any
	switch v := cur[last].(type) {
	case []any:
		list = v
	case nil:
		list = nil
	}
	for _, item := range list {
		if s, ok := item.(string); ok && s == value {
			return
		}
	}
	cur[last] = append(list, value)
}
