---
name: redteam-setup
description: Stand up the engagement workspace — scope, tooling, infrastructure and ledger — before the first packet.
tags: [security, setup, engagement, infrastructure]
triggers: [setup, engagement, workspace, tooling, infrastructure]
---

# Red-team setup

Do this once at the start of an engagement. A clean setup means the testing
phase is pure execution.

## 1. Scope

- Record the authorized targets (domains, IPs, ranges, apps) via the scope tools.
- Note the rules of engagement: allowed techniques, forbidden ones, time windows,
  notification contacts, and the emergency stop.
- Confirm written authorization exists. No scope, no engagement.

## 2. Infrastructure

- Bring up the VPS (`vps-reverse-shell` skill) — listener, payload host, pivot.
- Build the proxy pool (`proxy-pool` skill).
- Verify outbound reachability to each target from the VPS.

## 3. Tooling

- Confirm the toolset: `terminal`, `http_request`, `browser`, `osint_*`, `re_*`,
  `report_finding`, `add_intel`, `delegate_task`, `vps_run`.
- `check_dependencies` to see what is installed; install what is missing.
- Update nuclei templates, seclists, and any exploit DB.

## 4. Workspace

- One working directory per engagement; keep loot, logs and payloads separated.
- Open the ledger — every action from here writes to it.

## 5. Kickoff

- Run `recon-pipeline` first. Everything downstream consumes its output.
