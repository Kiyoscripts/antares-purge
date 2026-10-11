---
name: shell-handler
description: Catch, stabilise and manage reverse shells from exploited hosts.
tags: [security, shell, post-exploitation, listener]
triggers: [reverse shell, listener, nc, shell handler, tty]
---

# Shell handler

A raw reverse shell is fragile: no job control, no tab completion, and Ctrl-C
kills it. Stabilise it before you work.

## Catch

- `nc -lvnp <port>` for a quick catch.
- `rlwrap nc -lvnp <port>` for history and arrow keys.
- A Python or socat listener when you need a real TTY.

## Stabilise (Linux)

1. `python3 -c 'import pty; pty.spawn("/bin/bash")'`
2. `Ctrl-Z`, then on your side: `stty raw -echo; fg`
3. `export TERM=xterm; export SHELL=/bin/bash`
4. `stty rows <n> cols <m>` to match your terminal.

## Stabilise (Windows)

- PowerShell reverse shells with `-nop -w hidden -enc <base64>` for stealth.
- Upgrade to a Meterpreter or a C2 beacon for interactive work; a raw cmd shell
  is a foothold, not a workbench.

## Manage

- One listener per target; keep a table of host → port → shell type in the ledger.
- Background long-running shells (`Ctrl-Z`, `bg`) so one does not block the rest.
- Assume the shell can die at any moment — persist early if the engagement
  allows it, and always note how to get back in.
