#!/usr/bin/env python3
"""Self-test for scripts/download_stats_total.py (no network). Run from repo root."""

from __future__ import annotations

import os
import runpy
import sys
from pathlib import Path


def main() -> int:
    root = Path(__file__).resolve().parent.parent
    mod = runpy.run_path(str(root / "scripts" / "download_stats_total.py"))
    fmt = mod["_format_badge_total"]
    github_headers = mod["_github_api_headers"]
    sum_counts = mod["_sum_analytics_counts"]
    cases: list[tuple[int, str]] = [
        (999, "999"),
        (1000, "1k"),
        (1500, "1.5k"),
        (999_499, "999.5k"),
        (999_500, "1M"),
        (999_950, "1M"),
        (1_000_000, "1M"),
        (2_500_000, "2.5M"),
    ]
    for n, want in cases:
        got = fmt(n)
        if got != want:
            print(f"FAIL _format_badge_total({n}) = {got!r} want {want!r}", file=sys.stderr)
            return 1

    original_token = os.environ.pop("GITHUB_TOKEN", None)
    try:
        headers = github_headers()
        if "Authorization" in headers:
            print("FAIL expected no Authorization header without GITHUB_TOKEN", file=sys.stderr)
            return 1
        os.environ["GITHUB_TOKEN"] = "test-token"
        headers = github_headers()
        if headers.get("Authorization") != "Bearer test-token":
            print(f"FAIL expected bearer auth header, got {headers!r}", file=sys.stderr)
            return 1
    finally:
        if original_token is None:
            os.environ.pop("GITHUB_TOKEN", None)
        else:
            os.environ["GITHUB_TOKEN"] = original_token

    block_cases: list[tuple[dict[str, object], int]] = [
        ({"asc": 3, "asc --HEAD": 1}, 4),
        ({"asc": "3", "asc --HEAD": "1"}, 4),
        ({"asc": 3.0, "asc --HEAD": "1.0"}, 4),
        ({"asc": True, "other": None, "bad": "oops", "partial": 2.5}, 0),
    ]
    for block, want in block_cases:
        got = sum_counts(block)
        if got != want:
            print(f"FAIL _sum_analytics_counts({block!r}) = {got!r} want {want!r}", file=sys.stderr)
            return 1

    print(f"ok {len(cases)} _format_badge_total cases + auth/count parsing checks")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
