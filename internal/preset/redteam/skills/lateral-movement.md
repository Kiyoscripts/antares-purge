---
name: lateral-movement
description: Moving from a foothold to other authorized hosts — credential reuse, pivoting, and tunnelling.
tags: [security, lateral, pivot, post-exploitation]
triggers: [lateral movement, pivot, tunnel, post-exploitation, foothold]
---

# Lateral movement

**Scope gate first.** Every host you move to must be in the authorized set.
Moving to an out-of-scope host ends the engagement, not extends it.

## From the foothold

1. **Situational awareness.** Who am I (`whoami /all`, `id`), where am I
   (`hostname`, `ip a`), what can I reach (`netstat`, `arp -a`).
2. **Harvest local credentials.** Files (`.ssh/`, `.aws/credentials`, `.env`,
   config files), memory (mimikatz/`lsassy` on Windows), keyrings on Linux.
3. **Map the internal network.** ARP table, routing table, DNS servers, then a
   light scan of the discovered subnets with `terminal`.

## Pivot

- **SSH tunnelling** — `ssh -L`/`-R`/`-D` for a SOCKS proxy through the foothold.
- **chisel / frp / suochisel** — when SSH is not available; a client on the
  foothold, a server on your VPS (`vps_run`). See the tunnel skills.
- **Proxychains** — route tooling through the pivot:
  `proxychains4 nmap -sT -Pn <internal-host>`.

## Move

- **Credential reuse** — the same password on the next host. Try it first.
- **Pass-the-hash / pass-the-ticket** — impacket `psexec.py`, `wmiexec.py`,
  `smbexec.py` with a captured NTLM hash.
- **SSH keys** — a harvested private key that is authorized elsewhere.
- **Known CVEs** — unpatched internal services are often soft.

## Output

A graph: foothold → pivot → hosts reached → credentials used. Every hop to the
ledger. Note what each new host unlocks — that is the measure of the movement.
