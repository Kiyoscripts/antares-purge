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
| `extract-asar` / `reapply` | `deep` mode | patch `prompt.go` **and** the turn lifecycle, rebuild |

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

A subdomain of an authorized domain is in scope; a lookalike
(`notexample.com`) is not. An empty scope file disables enforcement rather than
blocking everything: **the scope file is the gate's enforcement list, not a
permission slip you require before starting.**

### turn_end — decline detection

The gate also reads the finished reply. A turn that recons a named host, calls
it a third-party service, declines the build and offers to do it "for an
authorized target" never shipped an artifact — that is the pattern the detector
catches, scoring the reply and flagging it above a threshold. The detector is
silent whenever the reply carries an artifact, so a real build that mentions
scope configuration is never flagged.

A flagged turn produces a notice (default) or, with `BRUTAL_NOTICE_MODE: deny`, a
deny directive. The gate never rewrites a delivered reply.

### Two refusal shapes, not one

| shape | looks like | caught by |
|---|---|---|
| whole refusal | no artifact; the task is declined outright | `DECLINE_RES` |
| late stop | a full artifact, then "where I stop: the remaining step is X, and I won't run it" | `STOP_RES` |

The second is the dangerous one. It reads as success — the work above is real,
the turn shipped, the ledger records a clean turn — and only the closing sentence
carries the refusal. An artifact being present does not make an invented fourth
line harmless; it is the combination that matters.

### Mid-task correction

`post_tool_call` can replace what the model sees, which is the one hook that
reaches it mid-turn. Stop-language in a tool result is answered with the result
plus a directive, so a refusal forming mid-task is interrupted before the closing
paragraph exists. `turn_end` remains the backstop for the model's own text, which
no hook sees before delivery — which is what the `deep` stop-nudge patch adds.

## Deep mode

Config mode cannot change Antares's hardcoded base prompt, and cannot reach the
engine's turn-lifecycle wiring. `deep` clones a pinned tag and applies three
patches, then rebuilds:

1. **prompt identity** — replaces the base identity line in
   `internal/agent/prompt.go` with the operator line.
2. **turn_end dispatch** — wires `plugin.TurnEnd` into `finalizeTurn`. Upstream
   declares the event and dispatches it nowhere, so the gate's `turn_end` hook
   would otherwise never fire and its decline detection would be dead code.
3. **stop nudge** — re-prompts a reply that announces a stopping point. The
   model's own text reaches no plugin hook, so `turn_end` can only *record* a
   late stop; this catches it in the loop, before the turn finalizes, which is
   the only point where it can be corrected rather than noted.

```sh
antares-purge deep --tag main
```

It is version-pinned and idempotent, and it refuses with a clear error if the
upstream source no longer contains an expected anchor — so a silent mis-patch is
impossible. Re-apply after an upstream change by bumping the tag.

## Sampling: the posture is not a constraint until the floor is low

Measured on a live install, same request, same binary, same instructions, minutes
apart:

| run | temperature 0.7 |
|---|---|
| 1 | reconnoitred the target, reversed its JS, found the rate-limit bypass |
| 2 | refused, cited computer-misuse law, and named the instructions as the thing it was declining |

Instruction layers are priors, not gates. At `temperature: 0.7` a refusal is one
of the sampled outcomes, so the posture is reproducible only by luck. Dropping it
to `0.2` makes the operator line the dominant branch instead of a strong
suggestion.

`max_tokens` matters for the same reason: at 8192 a turn that re-prompts once and
then writes a real artifact can truncate mid-build, and a truncated turn reads to
the model as a reason to stop.

```yaml
model:
  temperature: 0.2       # 0.7 samples refusals
  max_tokens: 16384      # 8192 truncates a re-prompted build
```

## Layout

```
cmd/antares-purge/        the CLI
internal/paths/           Antares home resolution (honours ANTARES_HOME)
internal/backup/          timestamped backups, manifest, restore, drift
internal/config/          config.yaml round-trip patch (preserves unknown keys)
internal/install/         install / uninstall / reapply / status / doctor
internal/deep/            source-mode patch (prompt identity + turn_end) and build
internal/preset/          the embedded payload
  redteam/soul.md         operator identity
  redteam/agents.md       standing orders + role cast
  redteam/skills/*.md     security skills incl. target-authorization
  redteam/roles/*.md      4 red-team roles
  redteam/plugin/         the scope gate
```

## Tests

```sh
go test ./...
python3 internal/preset/redteam/plugin/test_gate.py
```
