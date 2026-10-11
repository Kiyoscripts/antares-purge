---
name: intranet-scan
description: Fast internal network scanning with fscan/gogo once a pivot into the target network is established.
tags: [security, intranet, scanning, pivot, fscan]
triggers: [intranet, internal scan, fscan, gogo, internal network]
---

# Intranet scan

Once you have a pivot, the internal network is usually softer than the edge.
Scan it fast, then go deep on what matters.

## Establish the pivot first

No internal scan without a route in — an SSH tunnel, chisel SOCKS, or a shell on
an internal host. Confirm the internal ranges are in scope.

## Fast sweep (fscan)

`fscan -h 10.0.0.0/24 -p 22,80,443,445,3389,1433,3306,6379 -o fscan.txt`

fscan does port scan, service fingerprint, weak-credential checks and common
exploits in one pass. Run it on the pivot host itself for speed and to avoid
routing the whole sweep through your tunnel:

1. Upload fscan to the foothold (`vps_run` to host it, `curl` to pull).
2. Run against the discovered ranges.
3. Pull the output back into the ledger.

## Deeper sweep (gogo)

`gogo -i 10.0.0.0/24 -p common -o gogo.txt`

gogo is faster and quieter than fscan on large ranges, with better fingerprinting.
Use it for the wide sweep, fscan for the targeted exploit pass.

## Triage

- Prioritise: domain controllers, databases, CI/CD, jump hosts, backup servers.
- Every live host with a service → a lead. Weak creds → a finding.
- Feed the internal host list back into `lateral-movement`.
