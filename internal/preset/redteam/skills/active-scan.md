---
name: active-scan
description: Active port, service and vulnerability scanning against in-scope hosts, tuned to stay quiet.
tags: [security, scanning, nmap, nuclei]
triggers: [scan, port scan, nmap, nuclei, service detection]
---

# Active scan

**Scope gate first.** Active scanning sends packets to the target. Confirm the
host is authorized before the first scan.

## Ports and services

1. Fast pass: `nmap -T4 --top-ports 1000 -sV <target>` via `terminal`. Get the
   live set before going deep.
2. Full TCP on live hosts only: `nmap -p- -T4 --min-rate 1000 <target>`.
3. Service and version: `nmap -sV -sC -p <open> <target>`. The default scripts
   catch a surprising amount.
4. UDP where it matters: `nmap -sU --top-ports 50 <target>` — SNMP, DNS, NTP.

## Vulnerability

- `nuclei -u <url> -severity medium,high,critical` for known-issue coverage.
- Feed the fingerprints from `web-fingerprint` into targeted checks; a version
  match beats a generic template.
- Manual beats automated on auth, logic and access control — those need a human
  eye and a crafted request, not a template.

## Staying quiet

- Rate-limit (`--max-rate`, nuclei `-rl`) on production targets.
- Randomize scan order; avoid sequential sweeps that trip IDS.
- Scan from a proxy or VPS (`vps_run`) when source IP matters, so the engagement
  origin is not burned.

## Output

Open ports, service versions, and confirmed issues to the ledger. Mark anything
unconfirmed as a lead, not a finding — findings need proof.
