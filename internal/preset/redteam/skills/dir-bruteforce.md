---
name: dir-bruteforce
description: Discover hidden paths, files and endpoints on a web target through content brute forcing.
tags: [security, web, content-discovery, fuzzing]
triggers: [directory brute force, content discovery, gobuster, ffuf, fuzz]
---

# Directory brute force

Web roots lie about their size. Content discovery finds the admin panel, the
backup file, the API the frontend forgot to hide.

## Tools

- **ffuf** — fast, flexible, filters responses by size/words/code.
  `ffuf -u https://<target>/FUZZ -w /usr/share/seclists/Discovery/Web-Content/raft-medium-directories.txt -mc 200,204,301,302,401,403 -fs 0`
- **gobuster** — simple directory and vhost brute forcing.
- **feroxbuster** — recursive, follows discovered directories automatically.

## Wordlists

- `seclists/Discovery/Web-Content/` — directories, files, extensions.
- `raft-*` for balanced coverage; `directory-list-2.3-medium` for depth.
- Build a target-specific list from `web-fingerprint` findings and `osint_dorks`.

## Find more than directories

- **vhosts** — a host header wordlist reveals virtual hosts sharing the IP:
  `ffuf -u https://<ip>/ -H 'Host: FUZZ.<domain>' -w vhosts.txt -fs <default-size>`
- **files** — extensions matter: `.bak`, `.old`, `.zip`, `.git/config`,
  `.env`, `swagger.json`, `actuator/env`.
- **parameters** — `ffuf -u 'https://<target>/api?FUZZ=1' -w params.txt` finds
  hidden query parameters that change behaviour.

## Tuning

- Filter on response size and word count, not just status code — a soft-404
  returns 200 for everything.
- Rate-limit (`-rate`) on production; a 404 flood is loud.
- Record every hit with its status, size and a short note to the ledger.
