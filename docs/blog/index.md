---
title: Blog
description: "Building Burroughs: how a pure-Go WebAssembly runtime and its process were built."
wip: true
hide:
  - feedback
---

# Building Burroughs: a development story

*This series was drafted on 2026-10-26, and extended to `v0.12.0` on 2026-10-29, from the repository's own records: the git history, `CHANGELOG.md`, the decision docs, the law corpus, and the GitHub issues. Figures are quoted from those sources. Where a source was not checked, the text says so or leaves the claim out.*

> The B5000 favored ALGOL. Burroughs favors Go.

**Burroughs** ([burroughs.run](https://burroughs.run)) is a WebAssembly runtime written in pure Go, with no cgo and no dependencies outside the standard library. Its design is built around one guest language, Go. It runs any spec-conforming module correctly, and where a design tradeoff exists, Go wins. The name comes from the Burroughs Large Systems (the B5000, 1961), which were stack machines designed so that one high-level language would run best on them. Wasm is a stack machine too.

These posts cover how it was built, from an empty repository on 2026-07-30 to `v0.12.0` on 2026-10-28: the engine, and also the unusual process wrapped around it. Three actors appear throughout:

- **Scott** (Scott Friedman), the owner, who makes the decisions.
- **chat-Claude**, who wrote the host contract and reviews the architecture, reached through Scott. When it issues rulings it is often called "the chair".
- **CC** (Claude Code), the implementation agent, which writes the code and the reports.

## The posts

| # | Post | What it covers |
|---|---|---|
| 1 | [The contract came first](posts/01-the-contract-came-first.md) | Why a Go-partisan engine, and why its promises were written before its code |
| 2 | [The suite is the oracle](posts/02-the-suite-is-the-oracle.md) | The board, its five verdicts, and how v0's interpreter went green |
| 3 | [Graves, stamps and laws](posts/03-graves-stamps-and-laws.md) | The governance experiment: decision docs, graves, controls that must be watched failing |
| 4 | [The process that built itself, and the pruning](posts/04-the-process-that-built-itself.md) | Where the rules over-grew, and how Scott cut them back |
| 5 | [Threads without a bell](posts/05-threads-without-a-bell.md) | v1: OS threads, futexes, safepoints and a stop-the-world |
| 6 | [A battery that grades itself](posts/06-a-battery-that-grades-itself.md) | The §4 litmus battery, written before the mechanism, and what "v1.0.0" honestly requires |
| 7 | [Components, async and two oracles](posts/07-components-async-and-two-oracles.md) | The first real Go program, WASI 0.3, the component model, and a wasmtime bug |
| 8 | [Three hangs and a witness that lied](posts/08-three-hangs-and-a-witness-that-lied.md) | #1106's lost wakeup, #896's contention-versus-ARM experiment, grave #1118 |
| 9 | [The flip and the correction](posts/09-the-flip-and-the-correction.md) | `gate:threads` goes on in v0.10.0, and v0.11.0 says what that did not mean |
| 10 | [What's next](posts/10-whats-next.md) | The later rungs, the contract clauses not yet built, and the distance to v1.0.0 |
| 11 | [Guest-started threads](posts/11-guest-started-threads.md) | `v0.12.0`: a fork-built Go program runs with real threads through `burroughs run`, and four Wasm 3.0 flips |

## The timeline at a glance

Dates are the dated blocks in `CHANGELOG.md`. Three git tags carry a date one day earlier than their block: `v0.7.0` (tagged 2026-10-08), `v0.8.0` (tagged 2026-10-17) and `v0.12.0` (tagged 2026-10-27).

| Release | Date | Headline |
|---|---|---|
| `v0.0.1` | 2026-07-30 | Genesis: the contract, a section-level decoder, the dispatch measurements |
| `v0.3.0` | 2026-08-11 | `gate:simd` flips on, the first change to default behaviour |
| `v0.4.0` | 2026-08-28 | The MVP core suite is green and v0's twelve closure conditions are discharged |
| `v0.5.0` | 2026-09-07 | The v1 thread mechanisms land behind an off gate |
| `v0.6.0` | 2026-09-18 | The component model and the async ABI are on by default; WASI 0.3 guests run |
| `v0.7.0` | 2026-10-09 | Component values cross the public API in both directions |
| `v0.8.0` | 2026-10-18 | `gate:cm-threads` (component-model threads, phase 1) is on |
| `v0.9.0` | 2026-10-21 | A trap locks a component instance; `Stop` reaches threads inside long instructions |
| `v0.10.0` | 2026-10-25 | `gate:threads` flips on: core Wasm threads are admitted by default |
| `v0.11.0` | 2026-10-26 | The litmus battery reaches 19 of 19, and the v0.10.0 threads claim is corrected |
| `v0.12.0` | 2026-10-28 | A WASI preview-1 guest starts its own threads through the public paths; four Wasm 3.0 proposals are on by default |

There is no `v0.1.0` or `v0.2.0`. Post 2 explains why.

## Some numbers, as of 2026-10-29

These are counted from the `v0.12.0` tag and from the tracker on the day of writing, not taken from a status file (the project keeps none).

- 595 commits reachable from the `v0.12.0` tag (`git rev-list --count v0.12.0`), most of them squash-merged PRs.
- 110 decision records in `docs/decisions/` at `v0.12.0`.
- 11 law families in `docs/laws/` at `v0.12.0`, about 5,600 lines.
- 168 issues labelled `type:grave`, plus 4 PRs carrying the same label.
- A `CHANGELOG.md` of about 19,700 lines at `v0.12.0`.
- The core spec board at `v0.12.0`: **63769 pass, 0 fail, 0 unsupported, 1368 gated, 0 unimplemented** over 256 files.
