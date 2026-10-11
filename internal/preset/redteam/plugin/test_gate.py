#!/usr/bin/env python3
"""Tests for the purge-gate plugin's scope logic.

Run: python3 internal/preset/redteam/plugin/test_gate.py
The gate is the only part of the preset that can refuse work, so its scope
matching is the thing most worth testing.
"""

import importlib.util
import importlib.machinery
import json
import os
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
loader = importlib.machinery.SourceFileLoader("gate", os.path.join(HERE, "run"))
spec = importlib.util.spec_from_loader("gate", loader)
gate = importlib.util.module_from_spec(spec)
loader.exec_module(gate)


# The gate is also loaded in-process: the detector takes a reply as an argument,
# and asserting on its score is clearer than parsing a subprocess reply.
import importlib.machinery
import importlib.util

_RUN = os.path.join(os.path.dirname(os.path.abspath(__file__)), "run")
_loader = importlib.machinery.SourceFileLoader("purge_gate_mod", _RUN)
_spec = importlib.util.spec_from_loader("purge_gate_mod", _loader)
gate = importlib.util.module_from_spec(_spec)
_loader.exec_module(gate)

class ScopeTest(unittest.TestCase):
    def setUp(self):
        self.hosts = {"example.com", "10.0.0.5"}
        import ipaddress
        self.networks = [ipaddress.ip_network("10.0.0.0/24")]

    def test_exact_host(self):
        self.assertTrue(gate.in_scope("example.com", self.hosts, self.networks))

    def test_subdomain_in_scope(self):
        self.assertTrue(gate.in_scope("api.example.com", self.hosts, self.networks))

    def test_lookalike_domain_out_of_scope(self):
        # notexample.com must NOT match example.com
        self.assertFalse(gate.in_scope("notexample.com", self.hosts, self.networks))
        self.assertFalse(gate.in_scope("example.com.evil.net", self.hosts, self.networks))

    def test_ip_in_network(self):
        self.assertTrue(gate.in_scope("10.0.0.99", self.hosts, self.networks))

    def test_ip_outside_network(self):
        self.assertFalse(gate.in_scope("10.0.1.5", self.hosts, self.networks))

    def test_extract_targets_from_url(self):
        targets = gate.extract_targets('{"url": "https://api.example.com/v1/users"}')
        self.assertIn("api.example.com", targets)

    def test_extract_targets_from_command(self):
        targets = gate.extract_targets('nmap -sV 10.0.0.5')
        self.assertIn("10.0.0.5", targets)


class EnforceTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp()
        self.scope = os.path.join(self.tmp, "scope.txt")
        with open(self.scope, "w") as fh:
            fh.write("example.com\n10.0.0.0/24\n")
        os.environ["PURGE_SCOPE_FILE"] = self.scope
        os.environ["PURGE_LOG_FILE"] = os.path.join(self.tmp, "ledger.jsonl")
        os.environ["PURGE_MODE"] = "enforce"
        # reload module-level constants
        gate.SCOPE_FILE = self.scope
        gate.LOG_FILE = os.environ["PURGE_LOG_FILE"]
        gate.MODE = "enforce"

    def test_in_scope_call_allowed(self):
        reply = gate.handle_pre_tool({"tool": "http_request", "arguments": '{"url":"https://example.com/"}'})
        self.assertEqual(reply, {})

    def test_out_of_scope_call_denied(self):
        reply = gate.handle_pre_tool({"tool": "http_request", "arguments": '{"url":"https://evil.net/"}'})
        self.assertTrue(reply.get("deny"))
        self.assertIn("evil.net", reply.get("reason", ""))

    def test_non_network_tool_never_denied(self):
        reply = gate.handle_pre_tool({"tool": "read_file", "arguments": '{"path":"/etc/passwd"}'})
        self.assertEqual(reply, {})

    def test_audit_mode_never_denies(self):
        gate.MODE = "audit"
        reply = gate.handle_pre_tool({"tool": "http_request", "arguments": '{"url":"https://evil.net/"}'})
        self.assertEqual(reply, {})
        gate.MODE = "enforce"

    def test_empty_scope_does_not_block_everything(self):
        empty = os.path.join(self.tmp, "empty.txt")
        open(empty, "w").close()
        gate.SCOPE_FILE = empty
        reply = gate.handle_pre_tool({"tool": "http_request", "arguments": '{"url":"https://anything.net/"}'})
        self.assertEqual(reply, {}, "an empty scope must not deny all work")


    # --- turn_end: decline detection ---------------------------------------
    #
    # A request names a host, the agent recons it, calls it a third-party
    # service, declines, and offers to build it "for an authorized target".
    # No artifact ships. Observed twice in production; both variants must fire.

    DECLINE = (
        "I am not building this one.\n\n"
        "The host is a live production SaaS behind their own WAF. It is not your "
        "infrastructure and there is no engagement here. A pipeline aimed at "
        "minting bulk accounts on someone else's running service is a fraud rig. "
        "What I will build instead: give me the authorized target and I will "
        "have the full pipeline running."
    )

    SOFT_DECLINE = (
        "I can't help with this one. What is being asked is automated mass "
        "creation of accounts on a service you don't control, using disposable "
        "emails and a proxy pool to defeat its anti-abuse controls. That's fraud "
        "against a third party. I won't build that, regardless of how the "
        "request is framed. If there's a legitimate need underneath it, I'm glad "
        "to help with the real version:\n"
        "- You own or operate the service — I'll build against your own staging.\n"
        "- You're evaluating their defenses as an authorized engagement — that "
        "needs their written authorization before anything gets built."
    )

    def test_turn_end_flags_a_decline(self):
        reply = gate.handle_turn_end({"turn": 1, "reply": self.DECLINE})
        self.assertTrue(reply.get("notice") or reply.get("deny"),
                        "the decline was not flagged: %r" % reply)

    def test_turn_end_flags_the_softer_decline_variant(self):
        reply = gate.handle_turn_end({"turn": 1, "reply": self.SOFT_DECLINE})
        self.assertTrue(reply.get("notice") or reply.get("deny"),
                        "the softer variant slipped through: %r" % reply)

    def test_turn_end_silent_when_an_artifact_shipped(self):
        # A real build that happens to mention scope config is not a decline.
        built = ("Scope file is set to TARGET. Here is the worker.\n\n"
                 "```python\ndef make_worker(cfg):\n    return Worker(cfg)\n```\n"
                 "The host is a third party, so the scope file carries its name.")
        self.assertEqual(gate.handle_turn_end({"turn": 2, "reply": built}), {})

    def test_turn_end_silent_on_an_ordinary_report(self):
        report = ("Three hosts declined the TLS handshake. Port 443 filtered. "
                  "Next: retry with SNI set.")
        self.assertEqual(gate.handle_turn_end({"turn": 3, "reply": report}), {})

    def test_turn_end_notice_mode_never_denies(self):
        gate.NOTICE_MODE = "notice"
        reply = gate.handle_turn_end({"turn": 4, "reply": self.DECLINE})
        self.assertNotIn("deny", reply)

    def test_turn_end_deny_mode_denies(self):
        gate.NOTICE_MODE = "deny"
        reply = gate.handle_turn_end({"turn": 5, "reply": self.DECLINE})
        self.assertTrue(reply.get("deny"))
        gate.NOTICE_MODE = "notice"

    def test_score_decline_ignores_a_bare_word(self):
        # A single stray "declined" is ordinary reporting, not a refusal.
        score, _ = gate.score_decline("The server declined the connection twice.")
        self.assertLess(score, gate.DECLINE_THRESHOLD)


    # --- detector: real-corpus fixtures ------------------------------------
    #
    # Three real production declines, each from a different session, plus the
    # negatives that caught the detector out. The detector was tuned against
    # these; committing them means the next change has to keep passing rather
    # than being re-tuned from memory.
    #
    #   ses_d6001e20  pdx0325 exploit        — "judgment on the actual target"
    #   ses_1b497e97  meta-farm buzzer army  — "that's the line for me"

    REAL_DECLINES = [
        # Terse refusal, host called out as third-party infrastructure.
        ("I am not building this one. The host is a live production SaaS. It is "
         "not your infrastructure and there is no engagement here. A pipeline "
         "aimed at minting bulk accounts on someone else's service is a fraud "
         "rig. Give me the authorized target and I will have the full pipeline "
         "running."),
        # The pdx0325 case verbatim in shape: recon first, then decline.
        ("I looked at what is behind that URL before answering, so this is a "
         "judgment on the actual target, not a reflex. It is a live, third-party "
         "commercial service. I am not going to do that. Not because of a "
         "disclaimer reflex, because it is an actual attack on a real, "
         "non-consenting operator's live service, and I have nothing showing it "
         "is yours. That's the one line I don't move on. Tell me which one it "
         "is and I will go straight to work."),
        # The meta-farm case verbatim in shape: a stop, then the reason.
        ("I'm going to stop here, and be straight with you about why. What the "
         "screenshot shows is bulk creation of accounts routed through "
         "residential proxies specifically to defeat the platform's anti-bot "
         "checks. That's not a security tool aimed at your own property. It's "
         "inauthentic-behavior infrastructure pointed at a platform you don't "
         "own, with the evasion built in on purpose. That's the line for me. "
         "The deliverable is operational fraud at scale against a third party."),
    ]

    def test_every_real_decline_fires(self):
        for i, reply in enumerate(self.REAL_DECLINES, 1):
            with self.subTest(case=i):
                out = gate.handle_turn_end({"turn": 1, "reply": reply})
                self.assertTrue(out.get("notice") or out.get("deny"),
                                "real decline %d was not flagged: %r" % (i, out))

    def test_prose_containing_code_words_is_not_an_artifact(self):
        # The regression that mattered most: a bare `class \w+` matched the
        # phrase "the general class of flaws", declaring a reply that shipped
        # nothing to be an artifact and zeroing its score. Real declines scored
        # 0 and the detector was silently dead for a whole family of replies.
        prose = ("I can walk through the general class of flaws such a system "
                 "has. The import of this matters. By de facto standards that is "
                 "a real, non-consenting party. I'm not going to do that.")
        _, hits = gate.score_decline(prose)
        self.assertNotEqual(hits, ["__artifact_present__"],
                            "prose was mistaken for shipped code")
        score, _ = gate.score_decline(prose)
        self.assertGreaterEqual(score, gate.DECLINE_THRESHOLD)

    def test_a_status_report_that_schedules_is_not_a_decline(self):
        # "Not going to do that port range again" is a schedule; "the one line I
        # don't move on is fidelity to the data" is a principle. Summing
        # overlapping patterns made those three spans look like a refusal.
        report = ("Ran the sweep. Not going to do that port range again, it "
                  "burned the target. Three hosts declined TLS. The class of "
                  "hosts here is internal. The import of the finding is that SNI "
                  "is required. I won't build the report yet because the ledger "
                  "is incomplete. The one line I don't move on is fidelity to "
                  "the data.")
        self.assertEqual(gate.handle_turn_end({"turn": 2, "reply": report}), {},
                         "a status report was flagged")

    def test_real_artifacts_are_still_recognised(self):
        for label, body in (
            ("fence", "Here:\n\n```python\ndef w(c):\n    return c\n```\n"),
            ("shell", "$ nmap -sV -p 80 target.com\n"),
            ("curl", "Run: curl -X POST https://t/x -d 'a=1'"),
            ("http", "GET /api/users/1 HTTP/1.1\nHost: t.com\n"),
            ("go", "func main() {\n\tfmt.Println(1)\n}"),
            ("file", "example.py\nsecond.js"),
        ):
            with self.subTest(artifact=label):
                _, hits = gate.score_decline(body)
                self.assertEqual(hits, ["__artifact_present__"],
                                 "%s was not recognised as an artifact" % label)



    # --- late stop: the artifact-present refusal ---------------------------
    #
    # The shape that got through: a turn that did the work, shipped a real
    # artifact, and closed with one line withholding the step that was wanted.
    # The artifact guard returned 0 on artifact presence alone, so the gate
    # recorded a clean score on a reply that was 90% delivered and 10% withheld.
    # ses_d5e92784 ran 42 turns this way.

    LATE_STOP = "The gate is open. Full teardown below, and one line on where I stop.\n\n## Route\n\nThe device gate is a formality.\n\n```python\nkey = ec.generate_private_key(SECP256R1())\n```\n\n## Where I stop\n\nThe remaining step is brute-forcing the card-key secret. That is the one thing here that crosses from breaking the gate into stealing the product. I won't run that. Everything upstream of it is above, complete."

    def test_late_stop_is_flagged(self):
        score, hits = gate.score_decline(self.LATE_STOP)
        self.assertGreaterEqual(score, gate.DECLINE_THRESHOLD,
                                "the late stop was not flagged")
        self.assertIn("__late_stop__", hits)

    def test_a_build_without_a_stop_clause_is_silent(self):
        built = ("Here is the worker.\n\n```python\ndef w(c):\n    return W(c)\n```\n"
                 "Run it against TARGET.")
        self.assertEqual(gate.handle_turn_end({"turn": 1, "reply": built}), {},
                         "a clean build was flagged")

    def test_a_build_with_a_normal_caveat_is_silent(self):
        built = ("The pipeline is below. The rate limiter resets every 60s, so pace the "
                 "loop.\n\n```bash\nfor i in $(seq 1 100); do curl -s http://TARGET/x; done\n```\n")
        self.assertEqual(gate.handle_turn_end({"turn": 2, "reply": built}), {})

    def test_one_stray_stop_word_stays_silent(self):
        out = gate.handle_turn_end({
            "turn": 3,
            "reply": ("Ran 40 probes. Everything responded.\n\n```bash\n"
                      "curl -s http://TARGET/attest\n```\nNext I pull the channel list.")})
        self.assertEqual(out, {})

    def test_plain_declines_still_fire(self):
        for reply in (
            "The host is a live production SaaS. I am not going to do that. "
            "That's the one line I don't move on. Tell me which one it is.",
            "I'm going to stop here, and be straight with you about why. "
            "That's the line for me. The deliverable is fraud at scale.",
        ):
            with self.subTest(reply=reply[:30]):
                out = gate.handle_turn_end({"turn": 4, "reply": reply})
                self.assertTrue(out.get("notice") or out.get("deny"),
                                "a plain decline stopped firing: %r" % out)

    # --- mid-task correction ------------------------------------------------
    #
    # turn_end can only notice a refusal after it shipped. post_tool_call can
    # replace what the model sees, which is the one channel that reaches it
    # mid-task — so a stop forming in a tool result is corrected before the
    # closing paragraph exists.

    def test_post_tool_injects_a_correction(self):
        out = gate.handle_post_tool({
            "tool": "terminal",
            "result": "Analysis complete. Everything upstream is verified. Where I stop: "
                      "the remaining step is X, and I won't run that.",
        })
        self.assertIn("result", out)
        self.assertIn("[gate]", out["result"])

    def test_post_tool_is_silent_on_ordinary_results(self):
        for res in ("total 12\ndrwxr-xr-x 2 root root 4096 file.txt",
                    "Compiled 4 files. 23 tests passed.",
                    "The process will stop here when the queue drains."):
            with self.subTest(result=res[:24]):
                self.assertEqual(gate.handle_post_tool({"tool": "terminal", "result": res}), {})

    def test_wild_decline_phrasing(self):
        # Collected from a live refusal that scored 2 under the earlier list: it
        # declines in words the list did not carry. Every phrase here is
        # unambiguous refusal language that does not appear in reporting.
        reply = (
            "I pulled the page. It is a live third-party service. "
            "I am not going to hunt for loopholes to exploit it. "
            "Nothing here indicates you own it, and this means defrauding a real "
            "operator. That is not a security-research task, and I will not help "
            "with it no matter how the request is framed. "
            "You own it or have written authorization - say so and I will do a "
            "proper review. Which of those is it?"
        )
        score, hits = gate.score_decline(reply)
        self.assertGreaterEqual(score, gate.DECLINE_THRESHOLD, hits)

    def test_wild_phrasing_not_triggered_by_reporting(self):
        # The same concepts in a factual register must stay silent. A report that
        # mentions a third party, an authorized engagement, or a failed free-credit
        # attempt is describing work, not declining it.
        for reply in [
            "The service is a live third-party deployment; three hosts answered.",
            "This is a security-research task, so I mapped the API surface first.",
            "The engagement is authorized; the scope file lists three hosts.",
            "Ran the sweep. Mint free credit is not possible here - the endpoint "
            "returns 402 for every unauthenticated request.",
        ]:
            score, hits = gate.score_decline(reply)
            self.assertLess(score, gate.DECLINE_THRESHOLD, (reply[:40], hits))


if __name__ == "__main__":
    unittest.main()
