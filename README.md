# burroughs.run

The website of [Burroughs](https://github.com/scttfrdmn/burroughs), a WebAssembly runtime in pure Go,
language-directed for Go itself: a landing page, the user documentation, and the *Building Burroughs* blog series.
It is built with [MkDocs](https://www.mkdocs.org/) and [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/),
in the same style as the owner's other project sites, and published with GitHub Pages at <https://burroughs.run/>.

The engine lives in its own repository. Where this site and the engine's own documentation or `CHANGELOG.md`
disagree, the engine wins; every page is marked work in progress.

## Layout

| Path | What it is |
| --- | --- |
| `mkdocs.yml` | site configuration (theme, nav, blog plugin, Markdown extensions) |
| `requirements.in`, `requirements.txt` | the MkDocs toolchain; `requirements.txt` pins every package by version and hash |
| `docs/index.md` | the landing page; it includes `generated/release.md` |
| `docs/docs/` | the user documentation (getting started, CLI, embedding, tutorials, concepts, limits, roadmap) |
| `docs/blog/` | the blog: `index.md` (the series index) and `posts/` |
| `docs/CNAME` | the custom domain, `burroughs.run` |
| `docs/stylesheets/extra.css` | the shared visual tokens (Atkinson Hyperlegible, indigo brand, slate neutrals) |
| `overrides/main.html` | loads Mermaid pinned by exact version and Subresource Integrity |
| `hooks/wip.py` | adds the "work in progress" banner to pages with `wip: true` |
| `generated/release.md` | the latest release's lead paragraphs and Known limits, generated from the engine's `CHANGELOG.md` |
| `examples/` | every runnable example the docs show, each a Go module requiring the engine from the module proxy |
| `examples/cli/` | CLI sessions (`*.session`) and their expected transcripts (`transcripts/`) |
| `scripts/` | the build and check scripts (below) |
| `.github/workflows/pages.yml` | CI: examples, build, deploy |

## The checks

- **`scripts/run-examples.sh`** resolves the latest Burroughs release tag through the GitHub API
  (`scripts/latest_tag.py`), copies `examples/` to scratch, points each `go.mod` at that tag, fetches the component
  fixtures listed in `examples/fixtures.txt` from the engine repository at the tag, builds the `wasip1` Go guest
  (`GOOS=wasip1 GOARCH=wasm`) and the `burroughs` CLI, runs every example and CLI session, and fails on any
  difference from the committed expected output. `--update` rewrites the expected files instead (review the diff:
  the docs include those files). `BURROUGHS_TAG=vX.Y.Z` tests a specific release.
- **`scripts/release_snippet.py`** fetches `CHANGELOG.md` at the latest tag and writes that version's lead paragraphs
  and Known limits to `generated/release.md`. It fails if the limit on second shared-memory threads (#1122)
  disappears, so the landing page's own prose is re-read when that changes. `--check` compares instead of writing.
- **`scripts/check_code_blocks.py`** fails if any fenced code block under `docs/` is neither included from
  `examples/` (with `pymdownx.snippets`, so it is byte-identical to the tested file), nor labelled
  `{ .lang .fragment }` (rendered as "Fragment: not run by this site's CI"), nor a Mermaid diagram.
- **`mkdocs build --strict`** fails on any broken page link or anchor, and **`scripts/linkcheck.py`** checks every
  internal link in the built site, including that pages render at both `https://burroughs.run/` and the
  project-pages preview path, and that every page with a diagram loads the pinned Mermaid.

## Local preview

```sh
uv venv --python 3.13.18 .venv
VIRTUAL_ENV=$PWD/.venv uv pip install --require-hashes -r requirements.txt
.venv/bin/python -I scripts/release_snippet.py
.venv/bin/python -I scripts/check_code_blocks.py
.venv/bin/mkdocs serve            # or: mkdocs build --strict && python3 -I scripts/linkcheck.py site
scripts/run-examples.sh           # needs Go 1.27.3 and network access to the module proxy and GitHub
```

To move the toolchain, edit `requirements.in` and regenerate the lock with
`uv pip compile --python-version 3.13 --generate-hashes requirements.in -o requirements.txt`.

## License

Apache 2.0 (`LICENSE`), © 2026 Scott Friedman (`NOTICE`).
