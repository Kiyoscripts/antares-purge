---
name: passive-recon
description: Passive OSINT against a target without touching its infrastructure — registries, DNS, search indexes, breach data.
tags: [security, osint, recon, passive]
triggers: [passive recon, osint, whois, dns, breach, shodan]
---

# Passive recon

Nothing here sends a packet to the target. Safe to run before scope is fully
signed off, and safe to run against third parties that merely mention the target.

## What to collect

- **Registration** — `osint_whois`, `osint_domain`: registrant org, registrar,
  creation/expiry, nameservers. Nameserver reuse links sibling domains.
- **DNS** — `osint_dns`: A/AAAA/MX/TXT/NS/SOA. MX reveals mail provider; TXT
  reveals SPF includes (SaaS inventory), DKIM selectors, and sometimes internal
  hostnames in verification records.
- **Search** — `osint_dorks_live`, `osint_google`: indexed files (`filetype:`),
  login panels (`inurl:`), directory listings, error pages with stack traces.
- **Code** — `osint_github`: the org's repos and any leaked secrets, internal
  hostnames, or CI config that names infrastructure.
- **Exposure** — `osint_shodan`, `osint_ip`, `osint_reputation`: internet-facing
  services and their banners, without scanning.
- **People** — `osint_email`, `osint_email_full`, `osint_username`: naming
  convention for the org, which drives credential attacks later.

## Output

Every artifact to the ledger with its source. Note the naming convention, the
mail/cloud providers, and any hostname that looks internal — those seed the
active phase.
