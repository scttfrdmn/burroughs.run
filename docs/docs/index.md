---
wip: true
---

# Burroughs user documentation

*Written against `v0.11.0` (tagged 2026-10-26). Every example program, expected output and CLI transcript on these
pages is included from this site's `examples/` directory, which the site's CI runs against the latest Burroughs
release before anything publishes. A code block that is not run that way is labelled as a fragment.*

Burroughs is a WebAssembly runtime written in pure Go: no cgo, and no dependency outside the Go standard library. Its
module path is `github.com/scttfrdmn/burroughs`. It is **language-directed**: its host contract (threads,
preemption, the memory model at host boundaries, stacks, the WASI 0.3 event loop) is designed around what the Go
runtime needs from a host. Any guest that conforms to the spec should run correctly, and the upstream spec test suite
is how that is checked. Where a design choice has to favour one kind of guest, it favours Go.

The pipeline is decoder, then internal form, then validator, then interpreter. **There is no compiler**, and none is
on the published roadmap.

## Status at v0.11.0, in one screen

| Area | State |
| --- | --- |
| Core Wasm (MVP + Wasm 2.0 features) | Runs. The spec suite's core board has 0 failures (see [Conformance](concepts.md#conformance-the-spec-suite-board)). |
| SIMD, relaxed SIMD | On by default. |
| Threads **grammar** (shared memory, atomics, wait/notify) | On by default, on every platform (the release *claim* covers linux and darwin; the default itself does not check the OS). **No public path can start a second thread** yet ([#1122](https://github.com/scttfrdmn/burroughs/issues/1122)). |
| GC, exception handling, tail calls, memory64, multi-memory, extended-const | Gated **off**. The public API cannot turn them on. |
| WASI preview 1 (`wasip1` commands) | Public path: `WASIP1Config.Run` and `burroughs run`. |
| Component model, async ABI, component-model threads | On by default. Values that cross the boundary: `u32`, `string`, `list`, `record`. |
| WASI 0.2 host for components | Partial: `wasi:cli` and `wasi:io` streams, part of each. Filesystem, clocks, random, sockets and http are absent. |
| Platforms | 64-bit Linux and macOS are built and tested. Windows runs the tests in CI as a measurement, not a gate, and is unsupported. |
| Stack switching (contract §7) | Not started. |

Details and issue links are in [Known limits and status](limits-and-status.md).

## How these docs are organised

- **[Getting started](getting-started.md)**: install, build the CLI, and run your first module.
- **[CLI reference](cli.md)**: `burroughs run`, `inspect` and `version`, every flag, the value syntax, how a file
  is routed, and the exit codes.
- **[Embedding guide](embedding.md)**: the Go API, type by type, with what each call does and returns.
- **Tutorials**, each ending in a run:
  1. [Call an exported function from WAT](tutorials/01-call-an-export.md)
  2. [Run a Go program built for WASI preview 1](tutorials/02-wasip1-go-program.md)
  3. [Load a component and call typed exports](tutorials/03-components.md)
  4. [Traps, refusals, gates and the environment switches](tutorials/04-traps-and-refusals.md)
- **[Concepts](concepts.md)**: the contract in plain terms, gates, conformance, and the trap / refused / declined /
  unsupported / gated vocabulary.
- **[Known limits and status](limits-and-status.md)**: what is and is not supported, with issue numbers.
- **[Roadmap](roadmap.md)**: what is decided, without dates. Subject to change.

## The examples

Every example program lives in this site's repository under `examples/`. Each is its own Go module whose `go.mod`
requires `github.com/scttfrdmn/burroughs v0.11.0` from the Go module proxy, with no local `replace`. Before the site
publishes, its CI (`scripts/run-examples.sh`) resolves the latest Burroughs release tag, points each example at it,
runs it, and compares what it prints with the committed expected output; any difference fails the build. The
component fixtures that tutorial 3 and the CLI transcripts use are fetched from the engine repository's
`internal/component/testdata/` at that tag, and the `wasip1` Go guest is built with the stock toolchain
(`GOOS=wasip1 GOARCH=wasm`).

| Directory | What it shows |
| --- | --- |
| `examples/embed-core/` | `Instantiate`, `Call`, multiple results, argument checking, traps, `Global`, `ParseValue` |
| `examples/refusals/` | one module per error classification, and `BURROUGHS_THREADS=0` |
| `examples/wasip1/` | a Go `wasip1` guest (`guest/`) and a Go host that runs it (`host/`) |
| `examples/component/` | `LoadComponent`, record, string and list values, the trap lock, `Close`, `ComponentConfig.Run` |
| `examples/describe/` | tutorial 4's error-branching template, run over one error of each kind |
| `examples/cli/` | the `burroughs` CLI sessions whose transcripts these pages show, run with the CLI built at the tag |

## Where the authoritative text lives

These pages are a reader's guide. When they disagree with the source, the source wins:

- the package documentation (`go doc -all github.com/scttfrdmn/burroughs`);
- the repository's `README.md` and `CHANGELOG.md` (each release since `v0.7.0` has a **Known limits** list);
- `docs/burroughs-contract-v0.1.md`, the normative host contract;
- `docs/decisions/`, the decision records (ADRs) cited throughout, each with a `Status:` line.

Project state (issues, milestones, the roadmap) is kept on GitHub at
[scttfrdmn/burroughs](https://github.com/scttfrdmn/burroughs), not in files.

Burroughs is Apache 2.0, © 2026 Scott Friedman.
