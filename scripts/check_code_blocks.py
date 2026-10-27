"""Fail if any fenced code block on the site is runnable-looking but untested.

Every fenced block under docs/ must be exactly one of:

  * an include: its body is a single pymdownx.snippets line,
        --8<-- "examples/<path>"            (optionally with :section or :start:end)
    naming a file under examples/ (code, expected output, or a CLI transcript),
    all of which scripts/run-examples.sh runs or regenerates against the latest
    Burroughs release, or under generated/;
  * a fragment: its attributes carry the class `.fragment`, e.g.
        ``` { .go .fragment }
    which the site renders with the label "Fragment: not run by this site's CI";
  * a diagram: ```mermaid.

An included path must exist, and a section name must have its start and end
markers in the file. Usage: python3 scripts/check_code_blocks.py
"""

import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FENCE = re.compile(r"^(?P<indent>\s*)(?P<fence>`{3,}|~{3,})(?P<info>.*)$")
SNIPPET = re.compile(r'^\s*--8<--\s+"(?P<path>[^":]+)(?::(?P<rest>[^"]+))?"\s*$')


def blocks(path):
    lines = open(path, encoding="utf-8").read().split("\n")
    i = 0
    while i < len(lines):
        m = FENCE.match(lines[i])
        if not m:
            i += 1
            continue
        fence, indent, info, start = m.group("fence"), m.group("indent"), m.group("info").strip(), i + 1
        body = []
        i += 1
        while i < len(lines):
            close = re.match(r"^\s*" + re.escape(fence[0]) + "{" + str(len(fence)) + r",}\s*$", lines[i])
            if close:
                break
            body.append(lines[i])
            i += 1
        yield start, info, body
        i += 1


def classes(info):
    lang = info.split("{", 1)[0].strip()
    attrs = info[info.index("{") + 1 : info.rindex("}")] if "{" in info and "}" in info else ""
    cls = set(re.findall(r"\.([A-Za-z0-9_-]+)", attrs))
    if lang:
        cls.add(lang)
    return cls


def check_include(m, where):
    rel, rest = m.group("path"), m.group("rest")
    if not (rel.startswith("examples/") or rel.startswith("generated/")):
        return f"{where}: include {rel!r} is not under examples/ or generated/"
    full = os.path.join(ROOT, rel)
    if not os.path.isfile(full):
        return f"{where}: include {rel!r} does not exist"
    if rest and not re.fullmatch(r"-?\d+(:-?\d+)?", rest):
        text = open(full, encoding="utf-8").read()
        if f"--8<-- [start:{rest}]" not in text or f"--8<-- [end:{rest}]" not in text:
            return f"{where}: section {rest!r} has no start/end markers in {rel}"
    return None


def main():
    errors, counts = [], {"include": 0, "fragment": 0, "mermaid": 0}
    for base, _, files in os.walk(os.path.join(ROOT, "docs")):
        for name in sorted(files):
            if not name.endswith(".md"):
                continue
            path = os.path.join(base, name)
            rel = os.path.relpath(path, ROOT)
            for line, info, body in blocks(path):
                where = f"{rel}:{line}"
                cls = classes(info)
                content = [b for b in body if b.strip()]
                if "mermaid" in cls:
                    counts["mermaid"] += 1
                elif len(content) == 1 and SNIPPET.match(content[0]):
                    err = check_include(SNIPPET.match(content[0]), where)
                    if err:
                        errors.append(err)
                    else:
                        counts["include"] += 1
                elif "fragment" in cls:
                    counts["fragment"] += 1
                else:
                    errors.append(
                        f"{where}: code block ({info or 'no info string'}) is neither an include from "
                        "examples/, a .fragment, nor mermaid"
                    )
    for e in errors:
        print(e)
    print(
        f"check_code_blocks: {counts['include']} included from examples/ or generated/, "
        f"{counts['fragment']} fragments, {counts['mermaid']} diagrams, {len(errors)} errors"
    )
    sys.exit(1 if errors else 0)


if __name__ == "__main__":
    main()
