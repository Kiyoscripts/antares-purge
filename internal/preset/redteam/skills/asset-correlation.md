---
name: asset-correlation
description: Correlate hosts, certificates, ASNs and favicons into a single target asset map.
tags: [security, recon, correlation, assets]
triggers: [asset, correlation, asn, favicon, certificate, pivot]
---

# Asset correlation

Isolated hosts hide the shape of a target. Correlation finds the hosts nobody
listed, by linking what they share.

## Link by

- **Certificate** — one cert's SANs enumerate every hostname it covers. Search
  certificate transparency for the org and for each cert found.
- **Favicon hash** — the same app on many hosts shares a favicon hash. Hash each
  and search for siblings.
- **ASN** — every IP in the org's ASN is candidate infrastructure. Enumerate the
  netblock, then filter to live.
- **Nameserver** — sibling domains under the same NS are often the same org.
- **Analytics / tracking IDs** — a shared GA or tag ID ties disparate hosts to
  one owner.
- **TLS JARM** — a server fingerprint that clusters hosts running the same stack.

## Steps

1. Seed with the domains and IPs from `passive-recon`.
2. Expand: CT logs, favicon hashes, ASN netblocks, reverse-WHOIS.
3. Resolve and filter to the live set.
4. Cluster by shared attribute; each cluster is one owner's infrastructure.
5. Write the map to the ledger — hosts, clusters, and the attribute that links
   them. The map drives scope and the testing plan.
