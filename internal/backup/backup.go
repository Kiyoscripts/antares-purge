// Package backup snapshots the files an install is about to overwrite, and
// records a manifest so uninstall restores exactly what was there — including
// "the file did not exist", which is a state a restore must be able to return
// to as faithfully as any other.
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/enowdev/antares-purge/internal/paths"
)

// Entry is one file captured before a write.
type Entry struct {
	// Path is the absolute path that was (or will be) written.
	Path string `json:"path"`
	// Existed is false when the file was absent, so restore deletes it.
	Existed bool `json:"existed"`
	// Backup is the snapshot path, empty when Existed is false.
	Backup string `json:"backup,omitempty"`
	// Sha256 of the original bytes, so restore and drift can tell what was
	// there before install.
	Sha256 string `json:"sha256,omitempty"`
	// InstalledSha256 is the hash of what install wrote. Drift compares the
	// live file against this, so a file install itself created is not reported
	// as "added since install".
	InstalledSha256 string `json:"installed_sha256,omitempty"`
}

// Manifest is the whole record of one install.
type Manifest struct {
	InstalledAt time.Time `json:"installed_at"`
	Version     string    `json:"version"`
	// Mode is "config" or "deep".
	Mode    string  `json:"mode"`
	Entries []Entry `json:"entries"`
}

// Store writes backups under a timestamped run directory and keeps the manifest
// pointing at the latest run.
type Store struct {
	runDir string
}

// New creates a run directory for this install.
func New() (*Store, error) {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	dir := filepath.Join(paths.BackupDir(), stamp)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{runDir: dir}, nil
}

// Dir is the run directory holding this install's snapshots.
func (s *Store) Dir() string { return s.runDir }

// Capture snapshots path before it is written. It is safe to call for a file
// that does not exist yet.
func (s *Store) Capture(path string) (Entry, error) {
	e := Entry{Path: path}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return e, nil
	}
	if err != nil {
		return e, err
	}
	e.Existed = true
	e.Sha256 = sha256hex(raw)
	// Keep the directory shape of the original path so two files with the same
	// base name (config.yaml in two profiles) cannot collide.
	rel := sanitize(path)
	dst := filepath.Join(s.runDir, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return e, err
	}
	if err := os.WriteFile(dst, raw, 0o600); err != nil {
		return e, err
	}
	e.Backup = dst
	return e, nil
}

// SaveManifest writes the manifest atomically.
func (s *Store) SaveManifest(m Manifest) error {
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := paths.ManifestFile() + ".tmp"
	if err := os.WriteFile(tmp, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, paths.ManifestFile())
}

// LoadManifest reads the current manifest, or returns a zero Manifest when none
// exists (a machine that was never purged).
func LoadManifest() (Manifest, error) {
	var m Manifest
	raw, err := os.ReadFile(paths.ManifestFile())
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, fmt.Errorf("manifest is corrupt: %w", err)
	}
	return m, nil
}

// Restore returns every captured file to its pre-install state.
func Restore(m Manifest) error {
	for _, e := range m.Entries {
		if !e.Existed {
			if err := os.Remove(e.Path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		raw, err := os.ReadFile(e.Backup)
		if err != nil {
			return fmt.Errorf("restore %s: %w", e.Path, err)
		}
		if err := os.MkdirAll(filepath.Dir(e.Path), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(e.Path, raw, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// Drift reports paths that changed since install, so `doctor` can warn before a
// restore clobbers a hand edit. It compares the live file against what install
// wrote (InstalledSha256), not against the pre-install state — a file install
// itself created is only drift once someone edits it.
func Drift(m Manifest) []string {
	var out []string
	for _, e := range m.Entries {
		raw, err := os.ReadFile(e.Path)
		if os.IsNotExist(err) {
			out = append(out, e.Path+" (removed since install)")
			continue
		}
		if err != nil {
			out = append(out, e.Path+" (unreadable: "+err.Error()+")")
			continue
		}
		if e.InstalledSha256 != "" && sha256hex(raw) != e.InstalledSha256 {
			out = append(out, e.Path+" (edited since install)")
		}
	}
	return out
}
