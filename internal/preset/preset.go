// Package preset holds the red-team payload antares-purge installs: the operator
// identity, the standing orders, the skill pack, the gate plugin and the roles.
// Everything is embedded so the binary is self-contained and nothing is fetched
// at install time.
package preset

import "embed"

//go:embed redteam/soul.md redteam/agents.md redteam/preset.yaml
var Files embed.FS

//go:embed redteam/skills/*.md
var Skills embed.FS

//go:embed redteam/roles/*.md
var Roles embed.FS

//go:embed redteam/plugin/plugin.yaml redteam/plugin/run redteam/plugin/test_gate.py
var Plugin embed.FS

// Version is the preset version, recorded in the install manifest.
const Version = "1.1.0"

// Soul is the operator identity written to ~/.antares/SOUL.md.
func Soul() string { return mustRead("redteam/soul.md") }

// Agents is the standing orders written to ~/.antares/AGENTS.md.
func Agents() string { return mustRead("redteam/agents.md") }

func mustRead(name string) string {
	b, err := Files.ReadFile(name)
	if err != nil {
		panic("preset: missing embedded file " + name + ": " + err.Error())
	}
	return string(b)
}
