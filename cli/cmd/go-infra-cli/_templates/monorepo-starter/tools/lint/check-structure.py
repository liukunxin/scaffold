#!/usr/bin/env python3
"""Monorepo structure check: turn directory conventions into CI-blocking errors.

Usage (run from the repository root):

    python tools/lint/check-structure.py
    python tools/lint/check-structure.py /path/to/repo-root

Only checks that files / directories exist and that manifest files are present.
No code parsing. Exit code 0 = pass, 1 = violations found.

Console output is English on purpose: Chinese text shows up as mojibake on
Windows consoles using the legacy GBK code page (same reason as the CLI).
"""

from __future__ import annotations

import sys
from pathlib import Path

# 根目录必须存在的东西：少一个就说明这不是本骨架，或者被改坏了。
REQUIRED_ROOT_ENTRIES = [
    "go.work",
    "apps",
    "services",
    "packages",
    "contracts",
    "README.md",
]

# Go Project 必须是 single-starter 布局（见 .cursor/rules/00-architecture.mdc）。
# Dockerfile 是必备项：项目一律从 single-starter 模板生成（go-infra-cli mono add），
# 模板自带一份（Monorepo 内构建上下文 = 仓库根，见文件内注释）。
REQUIRED_GO_ENTRIES = [
    "go.mod",
    "cmd",
    "configs",
    "internal/app",
    "internal/bootstrap",
    "internal/infra",
    "internal/route",
    "Dockerfile",
    "README.md",
]

# 非 Go Project 的识别方式 -> 至少要有的文件。
MANIFESTS = ["go.mod", "package.json", "pyproject.toml"]


def _project_kind(target: Path) -> str | None:
    for manifest in MANIFESTS:
        if (target / manifest).is_file():
            return manifest
    return None


def _missing(target: Path, names: list[str]) -> list[str]:
    return [name for name in names if not (target / name).exists()]


def check(root: Path) -> int:
    failures: list[str] = []
    checked = 0

    missing_root = _missing(root, REQUIRED_ROOT_ENTRIES)
    if missing_root:
        failures.append(f"{root}: missing {', '.join(missing_root)}")
        print(f"root                  FAIL  missing {', '.join(missing_root)}")
    else:
        print("root                  ok")

    for category in ("apps", "services"):
        parent = root / category
        if not parent.is_dir():
            continue
        for target in sorted(p for p in parent.iterdir() if p.is_dir()):
            rel = f"{category}/{target.name}"
            kind = _project_kind(target)
            if kind is None:
                failures.append(f"{rel}: not a Project (missing {' / '.join(MANIFESTS)})")
                print(f"{rel:<21} FAIL  not a Project (no manifest file)")
                continue

            checked += 1
            required = REQUIRED_GO_ENTRIES if kind == "go.mod" else ["README.md"]
            missing = _missing(target, required)
            if missing:
                failures.append(f"{rel}: missing {', '.join(missing)}")
                print(f"{rel:<21} FAIL  missing {', '.join(missing)}")
            else:
                print(f"{rel:<21} ok    [{kind}]")

    # packages/<lang>/<name> 必须有清单文件，否则引用方无法解析。
    packages = root / "packages"
    if packages.is_dir():
        for lang in sorted(p for p in packages.iterdir() if p.is_dir()):
            for target in sorted(p for p in lang.iterdir() if p.is_dir()):
                rel = f"packages/{lang.name}/{target.name}"
                if _project_kind(target) is None:
                    failures.append(f"{rel}: missing manifest file")
                    print(f"{rel:<21} FAIL  missing manifest file")
                else:
                    checked += 1
                    print(f"{rel:<21} ok")

    print()
    if failures:
        print(f"structure check FAILED: {len(failures)} item(s) to fix (checked {checked} project(s))")
        for item in failures:
            print(f"  - {item}")
        return 1

    print(f"structure check passed (checked {checked} project(s))")
    return 0


def main(argv: list[str]) -> int:
    if len(argv) > 2:
        print(__doc__)
        return 2

    root = Path(argv[1]).resolve() if len(argv) == 2 else Path(__file__).resolve().parents[2]
    if not root.is_dir():
        print(f"path does not exist: {root}")
        return 2

    print(f"repo root: {root}\n")
    return check(root)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
