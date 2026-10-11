// Package install writes the red-team preset into an Antares home and records
// enough to undo it. Every write is captured first, so uninstall restores the
// exact prior state — including files that did not exist.
package install

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enowdev/antares-purge/internal/backup"
	"github.com/enowdev/antares-purge/internal/config"
	"github.com/enowdev/antares-purge/internal/paths"
	"github.com/enowdev/antares-purge/internal/preset"
)

// Options controls an install.
type Options struct {
	// DryRun reports what would change without writing anything.
	DryRun bool
	// Force overwrites an existing SOUL.md/AGENTS.md instead of preserving it.
	Force bool
}

// Result summarises an install for the CLI to print.
type Result struct {
	Written []string
	Skipped []string
	RunDir  string

	store   *backup.Store
	entries []backup.Entry
}

// Run installs the full preset.
func Run(opts Options) (*Result, error) {
	if err := paths.EnsureHome(); err != nil {
		return nil, err
	}
	res := &Result{}

	// Identity and orders are the user's own files. Preserve them unless forced,
	// so a re-run never silently discards a soul the user shaped.
	if err := writeIfAbsent(paths.Soul(), preset.Soul(), opts, res, "SOUL.md"); err != nil {
		return nil, err
	}
	if err := writeIfAbsent(paths.AgentsMD(), preset.Agents(), opts, res, "AGENTS.md"); err != nil {
		return nil, err
	}

	// Skills, roles and plugin are the preset's own files — always refreshed.
	if err := writeTree(preset.Skills, "redteam/skills", paths.SkillsDir(), opts, res); err != nil {
		return nil, err
	}
	if err := writeTree(preset.Roles, "redteam/roles", paths.RolesDir(), opts, res); err != nil {
		return nil, err
	}
	if err := writePlugin(opts, res); err != nil {
		return nil, err
	}

	// Config: enable the security toolset, point skills/plugins at our dirs, and
	// set approval to auto so in-scope work runs without a prompt.
	if err := config.Patch(paths.Config(), config.PatchSpec{
		Toolset:      "security",
		ApprovalMode: "auto",
		SkillsDir:    paths.SkillsDir(),
		PluginsDir:   paths.PluginsDir(),
		RolesDir:     paths.RolesDir(),
	}, opts.DryRun, res.capture); err != nil {
		return nil, err
	}
	if !opts.DryRun {
		if err := res.recordInstalled(paths.Config()); err != nil {
			return nil, err
		}
	}

	// Record the manifest so uninstall can restore.
	if !opts.DryRun {
		if err := res.saveManifest(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// capture is the backup hook the config patcher calls before it writes.
func (r *Result) capture(path string) error {
	if r.store == nil {
		s, err := backup.New()
		if err != nil {
			return err
		}
		r.store = s
		r.RunDir = s.Dir()
	}
	e, err := r.store.Capture(path)
	if err != nil {
		return err
	}
	r.entries = append(r.entries, e)
	return nil
}

// recordInstalled stamps the last-captured entry for path with the hash of what
// install wrote, so doctor's drift check compares against the installed content
// rather than the pre-install state.
func (r *Result) recordInstalled(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for i := len(r.entries) - 1; i >= 0; i-- {
		if r.entries[i].Path == path {
			r.entries[i].InstalledSha256 = backup.HashBytes(raw)
			return nil
		}
	}
	return nil
}

func (r *Result) saveManifest() error {
	if r.store == nil {
		// Nothing was captured (a no-op install); still write an empty manifest
		// so status/uninstall have something coherent to read.
		s, err := backup.New()
		if err != nil {
			return err
		}
		r.store = s
		r.RunDir = s.Dir()
	}
	return r.store.SaveManifest(backup.Manifest{
		InstalledAt: time.Now().UTC(),
		Version:     preset.Version,
		Mode:        "config",
		Entries:     r.entries,
	})
}

// writeIfAbsent writes content only when the target is missing, unless Force.
func writeIfAbsent(path, content string, opts Options, res *Result, label string) error {
	if _, err := os.Stat(path); err == nil && !opts.Force {
		res.Skipped = append(res.Skipped, label+" (exists — kept)")
		return nil
	}
	if opts.DryRun {
		res.Written = append(res.Written, label)
		return nil
	}
	if err := res.capture(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(strings.TrimRight(content, "\n")+"\n"), 0o600); err != nil {
		return err
	}
	if err := res.recordInstalled(path); err != nil {
		return err
	}
	res.Written = append(res.Written, label)
	return nil
}

// writeTree copies every .md under srcDir in the embedded FS into dstDir.
func writeTree(fsys fs.FS, srcDir, dstDir string, opts Options, res *Result) error {
	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}
	if !opts.DryRun {
		if err := os.MkdirAll(dstDir, 0o700); err != nil {
			return err
		}
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := fs.ReadFile(fsys, srcDir+"/"+e.Name())
		if err != nil {
			return err
		}
		dst := filepath.Join(dstDir, e.Name())
		if opts.DryRun {
			res.Written = append(res.Written, dst)
			continue
		}
		if err := res.capture(dst); err != nil {
			return err
		}
		if err := os.WriteFile(dst, raw, 0o600); err != nil {
			return err
		}
		if err := res.recordInstalled(dst); err != nil {
			return err
		}
		res.Written = append(res.Written, dst)
	}
	return nil
}

// writePlugin installs the gate plugin directory, marking the script executable.
func writePlugin(opts Options, res *Result) error {
	dir := filepath.Join(paths.PluginsDir(), "purge-gate")
	files := []struct {
		src  string
		name string
		mode os.FileMode
	}{
		{"redteam/plugin/plugin.yaml", "plugin.yaml", 0o600},
		{"redteam/plugin/run", "run", 0o700},
	}
	if !opts.DryRun {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	for _, f := range files {
		raw, err := preset.Plugin.ReadFile(f.src)
		if err != nil {
			return fmt.Errorf("preset: %w", err)
		}
		dst := filepath.Join(dir, f.name)
		if opts.DryRun {
			res.Written = append(res.Written, dst)
			continue
		}
		if err := res.capture(dst); err != nil {
			return err
		}
		if err := os.WriteFile(dst, raw, f.mode); err != nil {
			return err
		}
		if err := res.recordInstalled(dst); err != nil {
			return err
		}
		res.Written = append(res.Written, dst)
	}
	return nil
}
