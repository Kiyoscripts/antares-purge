// Package paths resolves the Antares state directory and the files antares-purge
// writes into it. It mirrors internal/config.Home so the tool and the agent
// always agree on where a file lives.
package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Home is the Antares state directory: ~/.antares, overridable with
// ANTARES_HOME (the same variable Antares itself honours).
func Home() string {
	if v := strings.TrimSpace(os.Getenv("ANTARES_HOME")); v != "" {
		return v
	}
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return ".antares"
	}
	return filepath.Join(h, ".antares")
}

// Path joins parts onto the Antares home directory.
func Path(parts ...string) string {
	return filepath.Join(append([]string{Home()}, parts...)...)
}

// The named files antares-purge touches.
func Soul() string       { return Path("SOUL.md") }
func AgentsMD() string   { return Path("AGENTS.md") }
func UserMD() string     { return Path("USER.md") }
func Config() string     { return Path("config.yaml") }
func SkillsDir() string  { return Path("skills") }
func PluginsDir() string { return Path("plugins") }
func RolesDir() string   { return Path("roles") }

// BackupDir is where antares-purge keeps its timestamped backups and manifest.
func BackupDir() string { return Path("purge-backups") }

// ManifestFile records what an install wrote, so uninstall can restore exactly.
func ManifestFile() string { return Path("purge-backups", "manifest.json") }

// EnsureHome creates the state tree Antares expects, so a fresh machine works
// even before Antares has ever run.
func EnsureHome() error {
	for _, d := range []string{"", "sessions", "logs", "skills", "uploads", "backups", "plugins", "tmp"} {
		if err := os.MkdirAll(Path(d), 0o700); err != nil {
			return err
		}
	}
	return nil
}
