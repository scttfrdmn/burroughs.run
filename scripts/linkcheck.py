"""Check every internal link in the built site.

For each .html file under the site directory, every href/src that is not an
external URL (http(s):, mailto:, data:, javascript:) must:

  * be relative (no leading "/"), so the site renders both at
    https://burroughs.run/ and at the project-pages preview
    https://scttfrdmn.github.io/burroughs.run/;
  * resolve to a file in the site (a directory resolves to its index.html);
  * if it has a #fragment, name an id (or <a name>) that exists in the target.

404.html is the one exception to "relative": GitHub Pages serves it at any
missing path, so MkDocs writes its links root-absolute from site_url. Its
links are resolved against the site root instead. (At the project-pages
preview URL, before the custom domain is live, that page's links point at the
domain root; every other page works at both.)

It also checks that every page containing a Mermaid diagram loads the pinned
mermaid script with its integrity hash, so Material's own unpinned fallback
(mermaid@11 from unpkg) never runs.

Usage: python3 scripts/linkcheck.py [site-dir]   (default: site)
"""

import html.parser
import os
import sys
import urllib.parse

MERMAID_SRI = 'integrity="sha384-EOXBFmc3gx5mb+vn0vPvvGqACToJD24hhacX5Yx+8NUUQrHIle/Qi5Bg9o3zKwW2"'
SKIP_SCHEMES = ("http:", "https:", "mailto:", "data:", "javascript:", "tel:")


class Page(html.parser.HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.links, self.ids = [], set()

    def handle_starttag(self, tag, attrs):
        a = dict(attrs)
        for key in ("id", "name") if tag == "a" else ("id",):
            if a.get(key):
                self.ids.add(a[key])
        for key in ("href", "src"):
            if a.get(key) is not None and not (tag == "link" and a.get("rel") in ("canonical",)):
                self.links.append((tag, key, a[key]))


def parse(path, cache):
    if path not in cache:
        p = Page()
        with open(path, encoding="utf-8", errors="replace") as f:
            p.feed(f.read())
        cache[path] = p
    return cache[path]


def main():
    site = os.path.abspath(sys.argv[1] if len(sys.argv) > 1 else "site")
    cache, errors, checked = {}, [], 0
    for base, _, files in os.walk(site):
        for name in files:
            if not name.endswith(".html"):
                continue
            src = os.path.join(base, name)
            rel_src = os.path.relpath(src, site)
            for tag, key, url in parse(src, cache).links:
                if url.startswith(SKIP_SCHEMES) or url.startswith("//"):
                    continue
                if tag == "link" and key == "href" and url == "":
                    continue
                checked += 1
                parts = urllib.parse.urlsplit(url)
                if parts.path.startswith("/") and rel_src == "404.html":
                    parts = parts._replace(path=os.path.relpath(site + parts.path, base) + ("/" if parts.path.endswith("/") else ""))
                elif parts.path.startswith("/"):
                    errors.append(f"{rel_src}: root-absolute link {url!r} breaks under a project-pages prefix")
                    continue
                target = os.path.normpath(os.path.join(base, urllib.parse.unquote(parts.path))) if parts.path else src
                if os.path.isdir(target):
                    target = os.path.join(target, "index.html")
                if not target.startswith(site + os.sep) and target != site:
                    errors.append(f"{rel_src}: link {url!r} leaves the site")
                    continue
                if not os.path.isfile(target):
                    errors.append(f"{rel_src}: link {url!r} -> missing {os.path.relpath(target, site)}")
                    continue
                if parts.fragment and target.endswith(".html"):
                    frag = urllib.parse.unquote(parts.fragment)
                    if frag not in parse(target, cache).ids:
                        errors.append(f"{rel_src}: link {url!r} -> no id {frag!r} in {os.path.relpath(target, site)}")
    for base, _, files in os.walk(site):
        for name in files:
            if name.endswith(".html"):
                text = open(os.path.join(base, name), encoding="utf-8", errors="replace").read()
                if 'class="mermaid"' in text and MERMAID_SRI not in text:
                    errors.append(f"{os.path.relpath(os.path.join(base, name), site)}: has a diagram but not the pinned mermaid script")
    for e in sorted(set(errors)):
        print(e)
    print(f"linkcheck: {checked} internal links checked, {len(set(errors))} broken")
    sys.exit(1 if errors else 0)


if __name__ == "__main__":
    main()
