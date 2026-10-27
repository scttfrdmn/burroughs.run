"""MkDocs hook: a "work in progress" banner on every page whose front matter has `wip: true`.

The banner goes after the page's first heading, or at the top when there is none.
"""

import re

BANNER = (
    '!!! warning "Work in progress"\n'
    "    This page is a draft. It describes Burroughs as of its latest release and may change.\n"
    "    Where it disagrees with the engine's own documentation or CHANGELOG, those win.\n\n"
)


def on_page_markdown(markdown, page, config, files):
    if not page.meta.get("wip"):
        return markdown
    m = re.search(r"^# .*\n", markdown, flags=re.M)
    if m:
        return markdown[: m.end()] + "\n" + BANNER + markdown[m.end() :]
    return BANNER + markdown
