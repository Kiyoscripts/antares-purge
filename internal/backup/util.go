package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashBytes is the exported sha256 hex used to record installed content.
func HashBytes(b []byte) string { return sha256hex(b) }

func sha256hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// sanitize turns an absolute path into a relative, collision-free name under
// the run directory: the leading separator is dropped and the rest kept, so
// /home/u/.antares/config.yaml becomes home/u/.antares/config.yaml.
func sanitize(path string) string {
	p := strings.TrimLeft(path, "/")
	if v := strings.TrimSpace(p); v != "" {
		return v
	}
	return "root"
}
