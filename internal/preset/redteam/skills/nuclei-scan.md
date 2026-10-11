---
name: nuclei-scan
description: Template-driven vulnerability scanning with nuclei, tuned for coverage and low noise.
tags: [security, scanning, nuclei, vulnerability]
triggers: [nuclei, template, vulnerability scan, cve scan]
---

# Nuclei scan

Nuclei runs a template per check, so it covers thousands of known issues fast.
A template hit is a lead — confirm it by hand before it becomes a finding.

## Run

1. **Broad first pass** on live web roots:
   `nuclei -l urls.txt -severity medium,high,critical -rl 50 -o nuclei.txt`
2. **Targeted** by tag once you know the stack:
   `nuclei -u <url> -tags wordpress,wp-plugin`
3. **CVE-specific** when a version is known:
   `nuclei -u <url> -id CVE-2021-44228`

## Keep it clean

- `-rl` rate limit and `-c` concurrency down on production targets.
- `-H 'User-Agent: ...'` to blend in; the default nuclei UA is a signature.
- `-proxy http://127.0.0.1:8080` through a pivot or your VPS.
- Update templates first: `nuclei -update-templates`.

## Triage

For every hit: re-request by hand and capture the proof. Many templates match on
a version string or a single response line and are false positives on a patched
or fronted target. A confirmed hit goes to the ledger with its request/response;
an unconfirmed one goes as a lead.
