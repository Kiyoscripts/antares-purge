---
name: web-fingerprint
description: Identify the stack, framework, server and version behind an HTTP service, and map its endpoints.
tags: [security, web, fingerprint, recon]
triggers: [fingerprint, stack, framework, wappalyzer, headers]
---

# Web fingerprint

Know what you are hitting before you hit it. Version and framework decide which
vulnerability class is worth your time.

## Steps

1. **Headers first.** `http_request` the root and read `Server`, `X-Powered-By`,
   `Set-Cookie` names, `Via`, CDN headers. Cookie names alone often name the
   framework (`PHPSESSID`, `JSESSIONID`, `laravel_session`, `connect.sid`).
2. **Error pages.** Request a path that 404s and one that 500s. Framework error
   pages name the framework and sometimes the version and debug state.
3. **Static asset paths.** `/static/`, `/assets/`, hashed bundle names, and
   source-map files (`*.js.map`) leak the build tool and sometimes source.
4. **Known paths.** Probe framework defaults: `/wp-login.php`, `/actuator`,
   `/api/`, `/.git/config`, `/server-status`, `/swagger-ui`.
5. **WAF detection.** Send a benign quote or angle bracket and watch for a block
   page; note the vendor so evasion can be chosen later.
6. **TLS.** Certificate SANs enumerate every hostname the cert covers — often the
   full subdomain set in one request.

## Output

Per endpoint: server, framework, version, WAF, auth mechanism. Feed the version
into the vulnerability phase; a version is a CVE lookup waiting to happen.
