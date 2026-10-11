---
name: credential-attack
description: Password spraying, credential stuffing and hash cracking against authorized authentication surfaces.
tags: [security, credentials, brute-force, spraying, hashcat]
triggers: [credential, password spray, brute force, hashcat, hydra, login]
---

# Credential attack

**Scope gate first.** Credential attacks lock accounts and trip alerts. Confirm
the auth surface is authorized and the rules of engagement allow it.

## Order of operations

1. **Enumerate before you spray.** A valid username list beats a big password
   list. Harvest from `osint_email`, breach data, and login error messages that
   differ for valid vs invalid users.
2. **Default credentials first.** Every device, appliance and CMS ships with
   them. One `admin:admin` beats a thousand guesses.
3. **Password spraying, not brute force.** One common password across many users
   avoids lockout and stays under the radar:
   `hydra -L users.txt -p 'Summer2025!' -t 4 -W 3 ssh://<host>`
   Rotate the password every window; respect the lockout threshold.
4. **Credential stuffing.** Reuse breach pairs (`osint_breach`) against the login.
   Watch for MFA on the response.
5. **Hash cracking.** Once you have hashes, identify with `hashid`/`hashcat
   --identify`, then:
   - `hashcat -m 1000 -a 0 ntlm.txt rockyou.txt` then rules
     (`-r best64.rule`).
   - Mask for known policy: `hashcat -m 1000 -a 3 ntlm.txt ?u?l?l?l?l?d?d?d`.
   - Kerberoast/AS-REP where AD: capture with impacket, crack offline.

## Evasion

- Slow down; spread across time windows. Lockouts are the tell.
- Vary source IP (`list_proxies`, `vps_run`) if the target rate-limits by IP.
- Match the client: a real browser `User-Agent` and TLS fingerprint via
  `http_request` for web logins.

## Output

Every valid credential to the ledger immediately, with the service and the
access it grants. Credentials are the pivot to the next phase.
