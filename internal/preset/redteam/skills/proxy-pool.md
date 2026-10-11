---
name: proxy-pool
description: Build and rotate a proxy pool for scanning and exploitation that must not burn a single origin IP.
tags: [security, proxy, opsec, scanning]
triggers: [proxy, pool, rotation, opsec, socks]
---

# Proxy pool

One IP doing all the work is one IP to ban. A rotating pool spreads the load and
hides the engagement origin.

## Sources

- **Residential / mobile** — best for evading IP reputation blocks on web
  targets; expensive, use for the critical requests only.
- **Datacenter** — cheap and fast for scanning; easily blocked by WAFs.
- **Your own VPS fleet** — one per target region; fully under your control.

## Build the pool

1. Collect proxies with their type (HTTP/SOCKS5), geo, and latency.
2. Health-check each: `curl -x <proxy> -s https://ifconfig.me` and time it.
3. Drop dead and slow ones; keep a scored list.
4. Register the pool with Antares via `list_proxies` so tools can draw from it.

## Rotate

- **Per request** for web fuzzing — a new IP each hit defeats rate limits.
- **Per host** for scanning — one IP per target keeps the scan coherent.
- **Sticky sessions** where the target needs a consistent session (auth flows).
- Rotate the `User-Agent` and TLS fingerprint alongside the IP; `http_request`
  presents a real browser fingerprint, so pair it with a matching proxy.

## Opsec

- Never mix the engagement origin with your own traffic.
- Log which proxy touched which target, for the report's timeline.
- Assume a proxy can be logging you; do not send plaintext secrets through one
  you do not own.
