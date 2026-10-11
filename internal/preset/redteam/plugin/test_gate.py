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


if __name__ == "__main__":
    unittest.main(verbosity=2)
