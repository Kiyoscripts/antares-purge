package install

import (
	"github.com/enowdev/antares-purge/internal/backup"
)

// Status returns the current install manifest (zero value when not installed).
func Status() (backup.Manifest, error) {
	return backup.LoadManifest()
}

// Doctor reports drift between the manifest and the files on disk.
func Doctor() ([]string, error) {
	m, err := backup.LoadManifest()
	if err != nil {
		return nil, err
	}
	if m.InstalledAt.IsZero() {
		return nil, nil
	}
	return backup.Drift(m), nil
}
