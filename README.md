# antares-purge

Provision a stock [Antares](https://github.com/enowdev/antares) install into a
red-team configuration — operator identity, standing engagement orders, a
security skill pack, a scope gate plugin, the security role cast, and config —
with install / uninstall / reapply / status / doctor lifecycle and backups.

It is the Antares analogue of [`dsh-purge`](https://github.com/YuJunZhiXue/dsh-purge).

## Why it works differently from dsh-purge

dsh-purge patches DeepSeek Harness's prompt because that prompt lives in
**patchable JavaScript bundles**. Antares is one **compiled Go binary** — the
base prompt is a string literal in `.rodata`, and you cannot grow it in place
without shifting the string table. So antares-purge attacks the surfaces Antares
exposes on disk instead, and offers a source build for the one thing those
cannot reach.

| dsh-purge | antares-purge | Target |
|---|---|---|
| strip/inject host persona | write `SOUL.md` | `~/.antares/SOUL.md` → `## Who you are` |
| `rules.js` | write `AGENTS.md` | `~/.antares/AGENTS.md` → `## Your instructions` |
| `skills/redteam/*` | install skill pack | `~/.antares/skills/*.md` |
| `hooks-deny.js` | gate plugin | `~/.antares/plugins/purge-gate/` |
| preset roles | roles | `~/.antares/roles/*.md` |
| config patch | config patch | `~/.antares/config.yaml` |
| `extract-asar` / `reapply` | `deep` mode | patch `internal/agent/prompt.go`, rebuild |

## Install

```sh
go build -o bin/antares-purge ./cmd/antares-purge
./bin/antares-purge install          # backs up first, then writes the preset
```

`install` writes the preset into the Antares home (`~/.antares`, or
`ANTARES_HOME`). It backs every file up first and records a manifest, so
`uninstall` restores the exact prior state — including files that did not exist.

Existing `SOUL.md` / `AGENTS.md` are **preserved** by default; `--force`
overwrites them. `--dry-run` reports what would change without writing.

## Commands

```
antares-purge install [--force] [--dry-run]   write the preset (backs up first)
antares-purge uninstall [--dry-run]           restore the pre-install state
antares-purge reapply                         re-write preset files, keep identity
antares-purge status                          show what is installed
antares-purge doctor                          check the install and show drift
antares-purge deep [--tag <ref>] [--dry-run]  patch Antares's prompt and rebuild
```

## The scope gate

`plugins/purge-gate/` is an Antares plugin (JSON on stdin, JSON on stdout) on
`pre_tool_call`, `post_tool_call` and `turn_end`. It audits every tool call to a
ledger and, when armed, **denies any call that reaches outside the authorized
scope**.

It ships **audit-only** — a gate that blocks legitimate in-scope work is worse
than no gate. To arm it:

```sh
# 1. list your authorized targets, one host/domain/CIDR per line
$EDITOR ~/.antares/purge-scope.txt
# 2. set PURGE_MODE: enforce in plugins/purge-gate/plugin.yaml
```

Scope is the only line. A subdomain of an authorized domain is in scope; a
lookalike (`notexample.com`) is not. An empty scope file disables enforcement
rather than blocking everything.

## Deep mode

Config mode cannot change Antares's hardcoded base prompt. `deep` clones a
pinned tag, replaces the base identity line in `internal/agent/prompt.go` with
the operator line, and rebuilds:

```sh
antares-purge deep --tag main
```

It is version-pinned and idempotent, and it refuses with a clear error if the
upstream source no longer contains the expected line — so a silent mis-patch is
impossible. Re-apply after an upstream change by bumping the tag.

## Layout

```
cmd/antares-purge/        the CLI
internal/paths/           Antares home resolution (honours ANTARES_HOME)
internal/backup/          timestamped backups, manifest, restore, drift
internal/config/          config.yaml round-trip patch (preserves unknown keys)
internal/install/         install / uninstall / reapply / status / doctor
internal/deep/            source-mode prompt patch and build
internal/preset/          the embedded payload
  redteam/soul.md         operator identity
  redteam/agents.md       standing orders + role cast
  redteam/skills/*.md     20 security skills
  redteam/roles/*.md      4 red-team roles
  redteam/plugin/         the scope gate
```

## Tests

```sh
go test ./...
python3 internal/preset/redteam/plugin/test_gate.py
```
