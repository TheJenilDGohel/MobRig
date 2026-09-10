"""
update_changelog.py — Automatically generates or updates CHANGELOG.md entries
from Conventional Commits between git tags.
"""

import argparse
import datetime
import os
import re
import subprocess
import sys

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
CHANGELOG_PATH = os.path.join(REPO_ROOT, "CHANGELOG.md")


def get_latest_tag() -> str:
    try:
        out = subprocess.check_output(
            ["git", "tag", "--list", "v*", "--sort=-version:refname"],
            cwd=REPO_ROOT,
            text=True,
            stderr=subprocess.DEVNULL,
        ).strip()
        tags = [t.strip() for t in out.splitlines() if t.strip()]
        return tags[0] if tags else ""
    except Exception:
        return ""


def get_commits_in_range(tag_range: str) -> list[str]:
    try:
        out = subprocess.check_output(
            ["git", "log", tag_range, "--format=%s"],
            cwd=REPO_ROOT,
            text=True,
            stderr=subprocess.DEVNULL,
        ).strip()
        return [c.strip() for c in out.splitlines() if c.strip()]
    except Exception:
        return []


def categorize_commits(commits: list[str]) -> dict[str, list[str]]:
    categories = {
        "Added": [],
        "Fixed": [],
        "Performance": [],
        "Documentation": [],
        "Changed": [],
    }

    for c in commits:
        # Ignore merge commits, release commits, or skip-ci commits
        if c.startswith("Merge ") or "[skip ci]" in c or c.startswith("docs(changelog):"):
            continue

        # Match conventional commits: type(scope): message OR type: message
        m = re.match(r"^([a-zA-Z]+)(?:\(([^)]+)\))?!?:\s*(.+)$", c)
        if m:
            ctype, scope, desc = m.group(1).lower(), m.group(2), m.group(3)
            formatted = f"**{scope}**: {desc}" if scope else desc
            # Capitalize first letter of description
            formatted = formatted[0].upper() + formatted[1:] if not scope else formatted

            if ctype == "feat":
                categories["Added"].append(formatted)
            elif ctype == "fix":
                categories["Fixed"].append(formatted)
            elif ctype == "perf":
                categories["Performance"].append(formatted)
            elif ctype == "docs":
                categories["Documentation"].append(formatted)
            else:
                categories["Changed"].append(formatted)
        else:
            # Non-conventional commit fallback
            categories["Changed"].append(c)

    return {k: v for k, v in categories.items() if v}


def build_version_section(version: str, date_str: str, categories: dict[str, list[str]]) -> str:
    clean_ver = version.lstrip("v")
    lines = [f"## [{clean_ver}] - {date_str}", ""]
    for cat, items in categories.items():
        lines.append(f"### {cat}")
        for it in items:
            lines.append(f"- {it}")
        lines.append("")
    return "\n".join(lines).rstrip() + "\n"


def update_changelog_file(version: str, date_str: str, section_md: str) -> None:
    if not os.path.exists(CHANGELOG_PATH):
        content = (
            "# Changelog\n\n"
            "All notable changes to this project will be documented in this file.\n\n"
            "The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n"
            "and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n"
        )
    else:
        with open(CHANGELOG_PATH, "r", encoding="utf-8") as f:
            content = f.read()

    clean_ver = version.lstrip("v")
    header_pattern = rf"## \[{re.escape(clean_ver)}\][^\n]*\n"
    
    # Check if this version section already exists
    if re.search(header_pattern, content):
        # Replace existing version section up to the next "## [" or end of file
        next_section = r"(?=\n## \[|\Z)"
        pattern = rf"(## \[{re.escape(clean_ver)}\][\s\S]*?){next_section}"
        new_content = re.sub(pattern, section_md.strip() + "\n", content, count=1)
    else:
        # Insert after the introduction (after the first "## [" or at end of header)
        match = re.search(r"\n(## \[)", content)
        if match:
            idx = match.start() + 1
            new_content = content[:idx] + section_md + "\n" + content[idx:]
        else:
            new_content = content.rstrip() + "\n\n" + section_md

    with open(CHANGELOG_PATH, "w", encoding="utf-8") as f:
        f.write(new_content)
    print(f"Updated {CHANGELOG_PATH} with version [{clean_ver}]")


def main():
    parser = argparse.ArgumentParser(description="Update CHANGELOG.md from git commits")
    parser.add_argument("--version", help="Release version (e.g. 0.1.1 or v0.1.1)")
    parser.add_argument("--date", help="Release date (YYYY-MM-DD)")
    parser.add_argument("--output-notes", help="File to write release notes markdown to")
    args = parser.parse_args()

    date_str = args.date or datetime.date.today().isoformat()
    latest_tag = get_latest_tag()

    version = args.version
    if not version:
        if not latest_tag:
            version = "0.1.0"
        else:
            # Default to latest_tag or increment patch
            version = latest_tag.lstrip("v")

    range_spec = f"{latest_tag}..HEAD" if latest_tag else "HEAD"
    commits = get_commits_in_range(range_spec)
    categories = categorize_commits(commits)

    # If no commits were found in range (or only unreleased), still format
    section_md = build_version_section(version, date_str, categories)

    if args.output_notes:
        with open(args.output_notes, "w", encoding="utf-8") as f:
            # Strip the ## [version] line for release body
            body_lines = section_md.splitlines()[2:]
            f.write("\n".join(body_lines).strip() + "\n")
        print(f"Wrote release notes to {args.output_notes}")

    update_changelog_file(version, date_str, section_md)


if __name__ == "__main__":
    main()
