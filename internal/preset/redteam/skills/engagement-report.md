---
name: engagement-report
description: Write the final assessment from the consolidated ledger — findings, evidence, impact, remediation.
tags: [security, report, findings, deliverable]
triggers: [report, write-up, findings, deliverable, assessment]
---

# Engagement report

The report is the deliverable. Every finding already carries its evidence; this
skill turns the ledger into a document the client can act on.

## Structure

1. **Executive summary** — what was tested, when, the scope, and the headline
   risk in plain language. One page. No jargon.
2. **Scope and methodology** — authorized targets, techniques used, limitations
   (what was out of scope, what blocked testing).
3. **Findings**, each with:
   - **Title** and **severity** (CVSS or a clear scale).
   - **Affected asset** — the exact host, service, endpoint.
   - **Evidence** — the request, the response, the screenshot, the artifact.
   - **Impact** — what an attacker gains, concretely.
   - **Remediation** — the specific fix, not "patch your systems".
4. **Attack narrative** — the chain from initial access to impact, if one was
   achieved. This is what shows the client the real risk.
5. **Appendix** — raw tool output, credential list (redacted as agreed), timeline.

## Rules

- A finding without evidence is not a finding. Drop it or mark it a lead.
- Severity reflects impact, not effort. A trivial-to-fix RCE is still critical.
- Remediation is specific and testable.
- Redact secrets per the rules of engagement; the client still needs the proof.
