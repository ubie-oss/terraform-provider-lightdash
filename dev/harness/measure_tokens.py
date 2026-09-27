#!/usr/bin/env python3
"""Count tokens in the repo-controlled agent instruction surface.

This is the telemetry this repo can actually record. Cursor, Claude Code,
Codex, and Gemini assemble the request elsewhere; their tool schemas, cache
hits, and per-request usage fields are not available here.

Run with a Python that has tiktoken installed:

    python3 -m venv /tmp/harness-tok
    /tmp/harness-tok/bin/pip install tiktoken
    /tmp/harness-tok/bin/python dev/harness/measure_tokens.py
    /tmp/harness-tok/bin/python dev/harness/measure_tokens.py --check
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HARNESS = Path(__file__).resolve().parent
PRICES_PATH = HARNESS / "prices.json"
FLAGS_PATH = HARNESS / "flags.json"

# Ratchet for the always-on prefix. Raise this only when a prompt edit is intentional.
ALWAYS_ON_BUDGET = 2050

# Files a rendered Cursor request included on every turn in the session that
# produced this harness (observed 2026-09-28). Glob rules are not in this set.
# Plain markdown the agents load on every turn. Rules with frontmatter are
# classified from alwaysApply / globs / description instead.
ALWAYS_ON_PATHS = (
    "AGENTS.md",
    "CLAUDE.md",
)


def repo_root() -> Path:
    return ROOT


def load_encoder():
    try:
        import tiktoken
    except ImportError:
        sys.stderr.write(
            "tiktoken is not installed. Example:\n"
            "  python3 -m venv /tmp/harness-tok\n"
            "  /tmp/harness-tok/bin/pip install tiktoken\n"
            "  /tmp/harness-tok/bin/python dev/harness/measure_tokens.py\n"
        )
        raise SystemExit(2)
    return tiktoken.get_encoding("o200k_base")


def parse_frontmatter(text: str) -> dict[str, str]:
    if not text.startswith("---\n"):
        return {}
    end = text.find("\n---", 4)
    if end == -1:
        return {}
    meta: dict[str, str] = {}
    for line in text[4:end].splitlines():
        if ":" not in line or line.startswith(" "):
            continue
        key, value = line.split(":", 1)
        meta[key.strip()] = value.strip()
    return meta


def classify_rule(path: Path, text: str) -> str:
    meta = parse_frontmatter(text)
    if meta.get("alwaysApply") == "true":
        return "always_on"
    globs = meta.get("globs", "")
    description = meta.get("description", "")
    if globs:
        return "glob_attached"
    if description:
        return "agent_requestable"
    return "manual_only"


def classify(path: Path, text: str) -> str:
    rel = path.relative_to(ROOT).as_posix()
    if rel in ALWAYS_ON_PATHS or rel == "AGENTS.md":
        return "always_on"
    if rel.startswith(".cursor/rules/") and rel.endswith(".mdc"):
        return classify_rule(path, text)
    if "/skills/" in rel:
        return "on_demand_skill"
    if rel.startswith(".claude/agents/"):
        return "subagent_body"
    return "unclassified"


def iter_instruction_files() -> list[Path]:
    found: list[Path] = []
    for rel in ("AGENTS.md", "CLAUDE.md"):
        path = ROOT / rel
        if path.exists():
            found.append(path)
    for folder in (ROOT / ".cursor" / "rules", ROOT / ".claude" / "agents"):
        if folder.exists():
            found.extend(sorted(p for p in folder.rglob("*") if p.is_file() and p.suffix in {".md", ".mdc"}))
    skills = ROOT / ".claude" / "skills"
    if skills.exists():
        found.extend(sorted(p for p in skills.rglob("*") if p.is_file() and p.suffix in {".md", ".go"}))
    return found


def count_tokens(encoder, text: str) -> int:
    return len(encoder.encode(text))


def file_record(encoder, path: Path, text: str | None = None) -> dict:
    body = path.read_text(encoding="utf-8") if text is None else text
    rel = path.relative_to(ROOT).as_posix()
    meta = parse_frontmatter(body)
    return {
        "path": rel,
        "class": classify(path, body),
        "tokens": count_tokens(encoder, body),
        "bytes": len(body.encode("utf-8")),
        "always_apply": meta.get("alwaysApply", ""),
        "globs": meta.get("globs", ""),
        "description_tokens": count_tokens(encoder, meta.get("description", "")) if meta.get("description") else 0,
    }


def load_prices() -> dict:
    return json.loads(PRICES_PATH.read_text(encoding="utf-8"))


def prefix_cost(tokens: int, model: dict, turns: int, cache_hit_rate: float) -> dict | None:
    """Cost of resending one prefix for a whole task.

    Turn 1 is uncached. Later turns are cached with probability cache_hit_rate.
    Output tokens are not part of the prefix and are omitted.
    """
    input_price = model.get("input_per_mtok")
    cached_price = model.get("cached_input_per_mtok")
    if input_price is None or cached_price is None:
        return None
    uncached_turns = 1 + (turns - 1) * (1 - cache_hit_rate)
    cached_turns = (turns - 1) * cache_hit_rate
    uncached_tokens = tokens * uncached_turns
    cached_tokens = tokens * cached_turns
    return {
        "uncached_input_tokens": round(uncached_tokens, 1),
        "cached_input_tokens": round(cached_tokens, 1),
        "output_tokens": 0,
        "usd": round((uncached_tokens * input_price + cached_tokens * cached_price) / 1_000_000, 6),
    }


def scenario_costs(tokens: int, prices: dict) -> dict:
    scenario = prices["illustrative_scenario"]
    turns = scenario["turns_per_task"]
    out = {"turns_per_task": turns, "note": scenario["note"], "models": {}}
    for name, model in prices["models"].items():
        out["models"][name] = {
            "no_cache": prefix_cost(tokens, model, turns, 0.0),
            "full_cache_after_first_turn": prefix_cost(tokens, model, turns, 1.0),
        }
    return out


def sum_class(records: list[dict], kind: str) -> int:
    return sum(r["tokens"] for r in records if r["class"] == kind)


def measure(encoder, overrides: dict[str, str] | None = None) -> dict:
    overrides = overrides or {}
    records = []
    for path in iter_instruction_files():
        rel = path.relative_to(ROOT).as_posix()
        records.append(file_record(encoder, path, overrides.get(rel)))
    records.sort(key=lambda r: r["path"])
    prices = load_prices()
    always = sum_class(records, "always_on")
    glob = sum_class(records, "glob_attached")
    return {
        "tokenizer": "tiktoken o200k_base",
        "per_request_usage_recorded": False,
        "always_on_tokens": always,
        "glob_attached_tokens": glob,
        "on_demand_skill_tokens": sum_class(records, "on_demand_skill"),
        "subagent_body_tokens": sum_class(records, "subagent_body"),
        "agent_requestable_tokens": sum_class(records, "agent_requestable"),
        "manual_only_tokens": sum_class(records, "manual_only"),
        "files": records,
        "illustrative_always_on_cost": scenario_costs(always, prices),
        "illustrative_provider_task_prefix_cost": scenario_costs(
            always + tokens_for(records, ".cursor/rules/provider-implementation.mdc") + tokens_for(records, ".cursor/rules/code-quality.mdc"),
            prices,
        ),
    }


def tokens_for(records: list[dict], rel: str) -> int:
    for record in records:
        if record["path"] == rel:
            return record["tokens"]
    return 0


def load_flags() -> dict:
    if not FLAGS_PATH.exists():
        return {}
    return json.loads(FLAGS_PATH.read_text(encoding="utf-8"))


def variant_overrides(flag: dict) -> dict[str, str]:
    overrides: dict[str, str] = {}
    for item in flag.get("replaces", []):
        src = HARNESS / item["src"]
        overrides[item["dest"]] = src.read_text(encoding="utf-8")
    return overrides


def print_human(report: dict) -> None:
    print(f"tokenizer: {report['tokenizer']}")
    print(f"per-request API usage recorded: {report['per_request_usage_recorded']}")
    print(f"always_on_tokens: {report['always_on_tokens']}")
    print(f"glob_attached_tokens: {report['glob_attached_tokens']}")
    print(f"on_demand_skill_tokens: {report['on_demand_skill_tokens']}")
    print(f"subagent_body_tokens: {report['subagent_body_tokens']}")
    print(f"manual_only_tokens: {report['manual_only_tokens']}")
    print("files:")
    for record in sorted(report["files"], key=lambda r: (-r["tokens"], r["path"])):
        if record["class"] == "on_demand_skill" and record["tokens"] < 400:
            continue
        print(f"  {record['tokens']:6d}  {record['class']:16s}  {record['path']}")
    provider = report["illustrative_provider_task_prefix_cost"]
    print("illustrative provider-task prefix, 8 turns, repo instructions only:")
    for name, costs in provider["models"].items():
        full = costs["full_cache_after_first_turn"]
        cold = costs["no_cache"]
        if full is None:
            print(f"  {name}: dollar rate unavailable")
            continue
        print(f"  {name}: ${full['usd']:.4f} if cached after turn 1; ${cold['usd']:.4f} if never cached")


def check_baseline(report: dict) -> int:
    current = report["always_on_tokens"]
    if current > ALWAYS_ON_BUDGET:
        sys.stderr.write(
            f"always-on instructions grew: {current} tokens > budget {ALWAYS_ON_BUDGET}\n"
        )
        return 1
    print(f"always-on tokens {current} <= budget {ALWAYS_ON_BUDGET}")
    return 0


def compare_variants(encoder) -> int:
    flags = load_flags()
    if not flags:
        sys.stderr.write(f"no flags at {FLAGS_PATH}\n")
        return 1
    base = measure(encoder)
    print(f"live always_on_tokens: {base['always_on_tokens']}")
    for name, flag in flags.items():
        if not isinstance(flag, dict) or "replaces" not in flag:
            continue
        overrides = variant_overrides(flag)
        report = measure(encoder, overrides)
        enabled = flag.get("enabled", False)
        parts = []
        for kind in ("always_on_tokens", "glob_attached_tokens", "subagent_body_tokens", "manual_only_tokens"):
            delta = report[kind] - base[kind]
            if delta:
                parts.append(f"{kind} {delta:+d}")
        print(f"{name}: enabled={enabled} " + (", ".join(parts) or "no token change"))
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    parser.add_argument("--variants", action="store_true")
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args()
    encoder = load_encoder()
    if args.variants:
        return compare_variants(encoder)
    report = measure(encoder)
    if args.json:
        print(json.dumps(report, indent=2))
    else:
        print_human(report)
    if args.check:
        return check_baseline(report)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
