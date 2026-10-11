---
name: recon-pipeline
description: End-to-end reconnaissance pipeline for an authorized target — passive first, then active, feeding the intel ledger.
tags: [security, recon, pipeline, enumeration]
triggers: [recon, reconnaissance, pipeline, attack surface, enumerate]
---

# Recon pipeline

You cannot test what you have not found. This pipeline runs passive first, then
active, and writes everything to the ledger as it goes.

**Scope gate.** Confirm the target is in the authorized set before the first
active step. Passive OSINT touches nothing and may run first.

## Passive (no packets to the target)

1. `osint_domain` and `osint_whois` — registrant, nameservers, creation date.
2. `osint_dns` — A/AAAA/MX/TXT/NS records; TXT often leaks SPF, cloud, and
   third-party SaaS.
3. `osint_dorks` / `osint_dorks_live` — indexed pages, exposed files, logins.
4. `osint_github` — org repos, leaked keys, internal hostnames in code.
5. `osint_breach` and `osint_email` — credential exposure for later spraying.
6. `osint_shodan` — internet-facing services the target may not know it exposes.

Record each result with `add_intel`. Passive recon is cheap; run it all.

## Active (in scope only)

1. Subdomain brute force and permutation — resolve, keep the live set.
2. Port and service scan with `terminal` (nmap or a lighter tool). Start with the
   top 1000, widen only on a live host.
3. `web-fingerprint` skill to identify stack, framework and version per endpoint.
4. Directory and content discovery on every web root found.
5. Screenshot every live HTTP service with `browser` — visual triage is fast.

## Output

A target map: hosts → ports → services → endpoints → stack. Every node in the
ledger via `report_finding`/`add_intel`, with the evidence (command, response,
screenshot path). The map is what the testing roles consume next.
