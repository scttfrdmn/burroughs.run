"""MkDocs hook: every blog post opens with a banner saying it is a dated record.

The chair's fifth ruling of 2026-10-28, item 5. The text is the ruling's:

    A dated record, written <date>. It is not updated as the project moves; the docs track the current release.

`<date>` is the post's `written:` front matter, the day the text was written. That is not `date:`, which the
blog plugin orders by and which is the date of the events the post describes. A post with no `written:` fails
the build rather than getting a date the hook made up.

The banner goes after the post's first heading, as the wip banner does. A post does not also carry `wip: true`:
that banner says a page "describes Burroughs as of its latest release", which is exactly what a dated record
does not do.
"""

import datetime
import re

from mkdocs.exceptions import PluginError


def banner(written: str) -> str:
    return (
        '!!! note "A dated record"\n'
        f"    A dated record, written {written}. It is not updated as the project moves; "
        "the docs track the current release.\n\n"
    )


def on_page_markdown(markdown, page, config, files):
    if not page.file.src_uri.startswith("blog/posts/"):
        return markdown
    written = page.meta.get("written")
    if written is None:
        raise PluginError(f"{page.file.src_uri}: a blog post needs `written:` front matter, the day it was written")
    if isinstance(written, (datetime.date, datetime.datetime)):
        written = written.strftime("%Y-%m-%d")
    if page.meta.get("wip"):
        raise PluginError(f"{page.file.src_uri}: a dated post does not carry `wip: true`; its banner says otherwise")
    m = re.search(r"^# .*\n", markdown, flags=re.M)
    if m:
        return markdown[: m.end()] + "\n" + banner(str(written)) + markdown[m.end() :]
    return banner(str(written)) + markdown
