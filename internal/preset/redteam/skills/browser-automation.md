---
name: browser-automation
description: Drive a real browser for authenticated testing, JS-heavy apps and session-based attacks.
tags: [security, browser, web, automation]
triggers: [browser, authenticated, javascript, session, spa]
---

# Browser automation

Many targets only reveal themselves to a real browser: SPAs, JS-gated routes,
session-bound flows. Antares drives Chrome via CDP through the `browser` tool.

## When to use it

- The app is a SPA and `http_request` returns an empty shell.
- You need an authenticated session to reach the interesting surface.
- The vulnerability is client-side: DOM XSS, postMessage, prototype pollution.
- You want a screenshot for the report's evidence.

## Workflow

1. **Navigate** to the target and let the app settle.
2. **Log in** with harvested credentials; the session persists in the browser
   profile, so subsequent requests are authenticated.
3. **Map the app** by walking the UI — the routes the frontend knows are the
   routes that exist, even if the sitemap does not list them.
4. **Capture** requests as you go; the network log is the API map.
5. **Test client-side** issues in the console: inject into DOM sinks, probe
   `postMessage` handlers, tamper with client-side auth checks.

## For the report

Screenshot each finding in the browser — a rendered proof beats a raw response
for the client. Note the URL, the action, and the session state.
