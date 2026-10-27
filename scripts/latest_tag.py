"""Print the latest release tag of scttfrdmn/burroughs, resolved through the GitHub API.

The repository publishes tags, not GitHub releases, so this lists tags
(GET /repos/{owner}/{repo}/tags) and picks the highest vMAJOR.MINOR.PATCH.
Pre-release and other tags are ignored. GITHUB_TOKEN is used when set, to
avoid the unauthenticated rate limit in CI.

Usage: python3 scripts/latest_tag.py [owner/repo]
"""

import json
import os
import re
import sys
import urllib.request

SEMVER = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def tags(repo):
    url = f"https://api.github.com/repos/{repo}/tags?per_page=100"
    while url:
        req = urllib.request.Request(url, headers={"Accept": "application/vnd.github+json"})
        token = os.environ.get("GITHUB_TOKEN")
        if token:
            req.add_header("Authorization", f"Bearer {token}")
        with urllib.request.urlopen(req, timeout=30) as resp:
            yield from (t["name"] for t in json.load(resp))
            link = resp.headers.get("Link", "")
        m = re.search(r'<([^>]+)>;\s*rel="next"', link)
        url = m.group(1) if m else None


def latest(repo):
    found = [(tuple(map(int, m.groups())), t) for t in tags(repo) if (m := SEMVER.match(t))]
    if not found:
        sys.exit(f"latest_tag: no vX.Y.Z tag found on {repo}")
    return max(found)[1]


if __name__ == "__main__":
    print(latest(sys.argv[1] if len(sys.argv) > 1 else "scttfrdmn/burroughs"))
