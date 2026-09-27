#!/usr/bin/env python3
"""Apply flagged instruction variants. Dry-run unless --apply is set.

Live files stay as they are while every flag in flags.json is false.
Rollback after --apply: git checkout -- <replaced paths>, then set enabled
back to false. Do not leave a flag true after reverting the files.
"""

from __future__ import annotations

import argparse
import json
import shutil
from pathlib import Path

HARNESS = Path(__file__).resolve().parent
ROOT = HARNESS.parents[1]
FLAGS_PATH = HARNESS / "flags.json"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apply", action="store_true", help="replace live instruction files")
    args = parser.parse_args()
    flags = json.loads(FLAGS_PATH.read_text(encoding="utf-8"))
    wrote = False
    for name, flag in flags.items():
        if not isinstance(flag, dict) or "replaces" not in flag:
            continue
        if not flag.get("enabled"):
            print(f"skip {name} (enabled=false)")
            continue
        for item in flag["replaces"]:
            src = HARNESS / item["src"]
            dest = ROOT / item["dest"]
            action = "write" if args.apply else "would write"
            print(f"{action} {dest.relative_to(ROOT)} from {src.relative_to(ROOT)}")
            if args.apply:
                dest.parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(src, dest)
                wrote = True
    if not args.apply:
        print("dry run. Re-run with --apply to replace live files.")
    elif not wrote:
        print("no enabled flags.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
