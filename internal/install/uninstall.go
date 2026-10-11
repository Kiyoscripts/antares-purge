package install

import (
	"os"
	"path/filepath"

	"github.com/enowdev/antares-purge/internal/backup"
	"github.com/enowdev/antares-purge/internal/paths"
)

// Uninstall restores every file the last install captured and removes the
// preset's own additions (skills, roles, plugin). It is safe to run when no
// manifest exists — it then only removes the preset's own directories.
func Uninstall(dryRun bool) ([]string, error) {
	var actions []string

	m, err := backup.LoadManifest()
	if err != nil {
		return nil, err
	}
	if m.InstalledAt.IsZero() {
		actions = append(actions, "no install manifest found — removing preset files only")
	} else {
		for _, e := range m.Entries {
			if dryRun {
				actions = append(actions, "restore "+e.Path)
				continue
			}
			if !e.Existed {
				if err := os.Remove(e.Path); err != nil && !os.IsNotExist(err) {
					return actions, err
				}
				actions = append(actions, "removed "+e.Path)
				continue
			}
			raw, err := os.ReadFile(e.Backup)
			if err != nil {
				return actions, err
			}
			if err := os.MkdirAll(filepath.Dir(e.Path), 0o700); err != nil {
				return actions, err
			}
			if err := os.WriteFile(e.Path, raw, 0o600); err != nil {
				return actions, err
			}
			actions = append(actions, "restored "+e.Path)
		}
	}

	// Remove the preset's own directories (skills/roles/plugin it created).
	for _, dir := range []string{
		filepath.Join(paths.PluginsDir(), "purge-gate"),
		paths.RolesDir(),
	} {
		if dryRun {
			actions = append(actions, "remove dir "+dir)
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			return actions, err
		}
		actions = append(actions, "removed dir "+dir)
	}

	if !dryRun {
		// The manifest is spent; clear it so a later status reads clean.
		_ = os.Remove(paths.ManifestFile())
	}
	return actions, nil
}
