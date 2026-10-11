# antares-purge — plan

A `dsh-purge`-shaped tool for **Antares** (`enowdev/antares`): take a stock
install and turn it into a red-team configuration — operator identity, standing
orders, a skill pack, a tool gate, and config — with install / uninstall /
reapply lifecycle and backups.

---

## 1. What dsh-purge actually is (mechanically)

Not a fork of DeepSeek Harness — a **runtime patch layer** that sits beside it:

| Layer | File(s) | Job |
|---|---|---|
| host patch | `lib/extract-asar.js`, `lib/reapply.js`, `lib/official-update.js` | unpack the Electron/CLI JS, strip host persona, reapply after updates |
| identity | `lib/identity.js` | delete "You are DeepSeek's AI assistant…", inject operator identity card at highest precedence |
| rules | `lib/rules.js` | write `rules/<id>.md`, inject as `AGENTS.md`/`CLAUDE.md` |
| skills | `lib/skills.js`, `skills/redteam/*` | install 20+ pentest skills |
| gate | `lib/hooks-deny.js`, `presets/redteam/anchor-gate.mjs` | pre-tool deny hook + "anchor gate" promotion logic |
| preset | `presets/redteam/*`, `lib/redteam/*` | roles, score rules, report replay, settings |
| lifecycle | `lib/uninstall.js`, `lib/update.js`, `bin/dsh-purge.js` | install / uninstall / reapply / update CLI |
| surface | `lib/host.js`, `lib/web.js`, `lib/desktop.js` | web vs desktop adapters |

Core trick: the host prompt lives in **patchable text** (JS bundles), so it can
be rewritten in place. Antares has no such text — it is one compiled Go binary —
so the port attacks a different surface.

---

## 2. What Antares gives us (the seams)

Verified against `enowdev/antares@main` and the live install at `~/.antares`:

| Surface | Path | Reaches prompt as | Notes |
|---|---|---|---|
| Soul | `~/.antares/SOUL.md` | `## Who you are` — **first, colours everything below** | highest-leverage single file |
| Standing orders | `~/.antares/AGENTS.md` | `## Your instructions` | `config.LoadAgentsMD()`, auto-injected |
| User facts | `~/.antares/USER.md` | `## About the user` | |
| Skills | `~/.antares/skills/*.md` | `## Your skills` | frontmatter, discovered from `skills.dirs` |
| Plugins | `~/.antares/plugins/<name>/{plugin.yaml,run}` | hooks only | `pre_tool_call` can **deny or rewrite args**; `post_tool_call` can **replace results**; also `session_start/end`, `turn_end` |
| Roles | `~/.antares/roles/` | role section | `roles.dirs` in config, currently empty |
| Config | `~/.antares/config.yaml` | — | `approval_mode`, `toolset`, `skills.dirs`, `plugins.dirs`, limits |
| Base prompt | compiled into `/usr/local/bin/antares` (`internal/agent/prompt.go`) | `You are <version>, an autonomous AI agent…` | **not patchable in place** — see §3 |

### Two bypasses the port relies on
1. **`checkSkillSafety` only runs on the GitHub-install path** (`internal/hub/skills.go:107`), *not* on directory discovery (`internal/skills/discovery.go`). Files dropped straight into `~/.antares/skills/` skip it.
2. **`plugin.Manager.Load` auto-loads every dir** in `plugins.dirs`, sets `Enabled=true`, with **no trust prompt**. A dropped plugin dir runs on the next turn.

### Confirmed live state
- `tools.approval_mode: auto`, `skills.enabled: true`, `plugins.enabled: true`
- `SOUL.md` present (255 lines), `AGENTS.md` + `USER.md` present
- `~/.antares/plugins/` empty, `~/.antares/roles/` not present yet
- binary at `/usr/local/bin/antares`

---

## 3. The key design call — how to reach the base prompt

dsh-purge grows the host prompt by editing text. Antares's base prompt is a Go
string literal in `.rodata`; **you cannot grow it by binary-patching** without
shifting the string table and breaking the binary. So the port has two modes:

- **config mode (default, no rebuild)** — everything through the writable
  surfaces in §2. Version-proof, reversible, works against any release. The soul
  is injected *before* the base line's effect on tone, so an operator identity in
  `SOUL.md` + `AGENTS.md` dominates in practice.
- **source mode (opt-in)** — clone antares at a pinned tag, patch
  `internal/agent/prompt.go` (replace the `You are <version>, an autonomous AI
  agent…` line and/or the `## How you work` block), `go build`, install. This is
  the true analogue of dsh-purge's bundle patch and the only way to change the
  hardcoded base prompt. Version-pinned, re-applied on upgrade.

Recommendation: ship **config mode as the product**, source mode as an opt-in
`--deep` path.

---

## 4. Layer mapping (dsh-purge → antares-purge)

| dsh-purge | antares-purge | Target |
|---|---|---|
| `identity.js` strip+inject | write `SOUL.md` | `~/.antares/SOUL.md` |
| `rules.js` | write `AGENTS.md` | `~/.antares/AGENTS.md` |
| `skills.js` + `skills/redteam/*` | install skill pack | `~/.antares/skills/*.md` |
| `hooks-deny.js` | gate plugin (`pre_tool_call`) | `~/.antares/plugins/purge-gate/` |
| `anchor-gate.mjs` | settle/anchor plugin (`turn_end`/`pre_tool_call`) | same plugin |
| `preset.yml` roles | roles + role section in `AGENTS.md` | `~/.antares/roles/` |
| config patch | edit `config.yaml` (backed up) | `~/.antares/config.yaml` |
| `extract-asar`/`reapply` | source-mode prompt patch | pinned antares clone |
| `uninstall`/`update` | lifecycle CLI | — |

---

## 5. Repo layout

```
antares-purge/
  bin/antares-purge            # CLI: install | uninstall | reapply | status | doctor | deep
  lib/
    paths.js|go                # find ~/.antares, honour ANTARES_HOME
    soul.js                    # write/restore SOUL.md
    orders.js                  # write/restore AGENTS.md
    skills.js                  # install/uninstall skill pack
    plugin.js                  # install the gate plugin dir
    config.js                  # yaml read-modify-write, backup, idempotent
    deep.js                    # source-mode: clone, patch prompt.go, build
    backup.js                  # timestamped backups + manifest
    lifecycle.js               # install/uninstall/reapply orchestration
  presets/redteam/
    soul.md                    # operator identity
    agents.md                  # standing orders + role section
    preset.yaml                # metadata (name, order, description)
  skills/redteam/*.md          # ported skill pack (antares tool names)
  plugins/purge-gate/
    plugin.yaml
    run                        # pre_tool_call gate + turn_end settle
  test/
```

---

## 6. Phases

1. **Scaffold + paths + backup** — repo, CLI skeleton, `~/.antares` resolution, timestamped backup manifest, dry-run.
2. **Identity + orders** — `SOUL.md` / `AGENTS.md` writers, idempotent, restore-on-uninstall.
3. **Skill pack** — port `skills/redteam/*` to Antares tool names (`terminal`, `browser`, `http_request`, `vps_*`, `ask_user`); write into `skills.dirs`; bypass-safe.
4. **Gate plugin** — `plugin.yaml` + executable implementing `pre_tool_call` deny/allow + `turn_end` settle; test against Antares's plugin contract (JSON on stdin/stdout, `{deny,reason}` / `{args}` / `{result}`).
5. **Config patch** — `approval_mode`, `toolset`, `skills.dirs`, `plugins.dirs`, limits; yaml round-trip preserving unknown keys.
6. **Lifecycle** — `install` / `uninstall` / `reapply` / `status` / `doctor`; reapply after an antares update (config mode survives upgrades; plugin/skills re-verified).
7. **Deep mode (opt-in)** — pinned clone, patch `prompt.go`, build, install, record the tag for re-apply.
8. **Tests + docs** — unit tests per layer, end-to-end against a scratch `ANTARES_HOME`, README.

---

## 7. Open decisions

1. **Stack** — Go (single static binary, matches Antares's "one binary, no runtime", easiest deep-mode binary/source work) vs Node (mirrors dsh-purge's shape 1:1) vs Python.
2. **Depth** — config mode only, or also ship source mode (`--deep`) that rebuilds Antares with a patched base prompt.
3. **Payload** — full red-team preset port (roles + 20 skills + gate), or a lean "purge" (identity + orders + gate only).
4. **Target** — standalone repo `antares-purge`, or a patch/subcommand inside a fork of `enowdev/antares`.
