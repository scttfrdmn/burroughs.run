---
wip: true
---

# Known limits and status

*This list is as of `v0.12.0` (2026-10-28). It is drawn from that release's **Known limits** in
`CHANGELOG.md`, from the repository README, and from behaviour observed while writing these docs; observed items are
marked. The issue states were read from GitHub on 2026-10-28. For anything later, check the newest release block in
`CHANGELOG.md`.*

## Phase

Burroughs is in **v1, the threads phase** (decision 0099). v0's closure conditions were all discharged at `v0.4.0`.
v1 is not finished, and `1.0.0` is reserved: the project stays on `0.x`. Releases map minors to milestones, so a
version number says which conformance milestone was reached.

## What works

- **Core WebAssembly**, MVP and Wasm 2.0, plus SIMD, relaxed SIMD, and since `v0.12.0` extended constant
  expressions, tail calls, multi-memory and memory64, through `Instantiate` / `Instance.Call` and `burroughs run`. The
  spec suite's core board reported 0 failures at `v0.12.0` (see
  [Conformance](concepts.md#conformance-the-spec-suite-board)).
- **WASI preview 1 commands**, such as Go programs built with `GOOS=wasip1`, through `WASIP1Config.Run` and
  `burroughs run`, with args, env, the standard streams, read-only preopens and writable scratch preopens.
- **Guest-started threads on that path** (since `v0.12.0`, [#1122](https://github.com/scttfrdmn/burroughs/issues/1122)).
  A module that imports `"burroughs" "spawn"`, with type `(i32) -> ()` or `(i32) -> (i32)`, and exports `mstart`
  of type `(i32) -> ()`, starts a new agent on an OS thread over its shared memory each time it calls the import.
  Contract §2 T-1 names the door. The release's end-to-end witness is a fork-built Go program whose output matches
  stock Go's, with several agents executing at once.
- **Components**: loading, typed calls with `u32` / `string` / `list` / `record` values, the async component ABI,
  cancellation through `context`, `Close`, and `wasi:cli/run` worlds.
- **Component-model threads** (`gate:cm-threads`, phase 1) start through the component paths. They are cooperative
  and run one at a time per instance.

## Limits on default-on paths (from v0.12.0's release notes)

- **A guest that calls `runtime.LockOSThread` and then needs another M crashes on the Phase 4 fork**
  ([#1157](https://github.com/scttfrdmn/burroughs/issues/1157), open), with *"fatal error: on a locked thread with no
  template thread"*. The throw is the fork runtime's own, not the engine's, and the fix is the fork's.
- **Threads start only where a guest asks.** The `"burroughs" "spawn"` door is served on the WASI preview-1 paths
  (`WASIP1Config.Run`, `burroughs run`). The rest is unchanged from `v0.11.0`:
  - core `Instantiate` supplies no imports, so a core module's shared memory has one agent, and
    `memory.atomic.wait` on it can only time out (observed: [tutorial 4](tutorials/04-traps-and-refusals.md));
  - the public API has no host-side spawn, join, stop or host-function registration (host-started threads are not
    built);
  - a component's `thread.spawn-ref` is refused by name, and `thread.spawn-indirect` is refused at decode.
- **A shared-everything-threads table is refused as "malformed", not by name**
  ([#1120](https://github.com/scttfrdmn/burroughs/issues/1120), open). It is still a refusal, but the category is
  wrong.
- **`Close` with a zero bound returns `nil` on Windows** where `ErrCloseIncomplete` is wanted
  ([#945](https://github.com/scttfrdmn/burroughs/issues/945), open). Windows is outside the support claim.
- **A Go guest's stop-the-world still depends on the fork compiler's inserted loop checks**
  ([#1064](https://github.com/scttfrdmn/burroughs/issues/1064), open, parked). Contract §3 SP-1 says the guest
  should not need to instrument itself.
- **A Go runtime fatal was seen once on CI's arm64 runner** in `internal/spec`
  ([#1089](https://github.com/scttfrdmn/burroughs/issues/1089), open). It has not recurred.
- **`array.new*`'s backing allocation is not polled**, so a stop can take more than 1 ms on a GC array of roughly 64
  Ki elements or more (ADR 0104). This only matters with GC, which is gated off by default and cannot be turned on
  publicly.

## Limits of the public surface

- **No host imports on the core path.** `Config` has no import set and no feature set. Any import is
  `ErrUnlinkable` at load. The only host APIs a guest can reach are WASI preview 1 (via `WASIP1Config`) and the WASI
  0.2 subset for components.
- **Two proposals cannot be enabled at all through the public API**: GC and exception handling. The only `Feature`
  is `threads`. (The four Wasm 3.0 proposals that flipped in `v0.12.0` are on by default, and can only be turned
  *off*, by their environment rollbacks.)
- **Core calls cannot be cancelled.** `Instance.Call` takes no `context`.
- **No exported-global enumeration** and no global writes. Memories and tables are not exposed.
- **Component values**: only `u32`, `string`, `list` and `record` cross the boundary (a `tuple` arrives as a record).
  `bool`, the other integer and float widths, `char`, `variant`, `enum`, `option`, `result`, `flags` and resource
  handles are refused by name as `ErrUnsupported`.
- **Calls into one `Component` execute one at a time** (ADR 0101), as the component model specifies.
- **`Instance` concurrency is not documented.**

## WASI coverage

- **Preview 1**: reads under granted directories, the standard streams, args, env and `proc_exit`. Under a
  `Writable` grant, `path_create_directory`, `path_unlink_file`, `path_remove_directory`, `fd_pwrite` and
  `path_open`'s create and truncate modes also work. Twenty functions still answer `ENOSYS` (ADR 0091's amendment counts 21; the host refuses 20).
- **WASI 0.2 for components is partial.**
  - `wasi:cli`: run, exit, stdout, stderr and stdin. `environment.get-arguments` returns `ComponentConfig.Args`.
    **`get-environment` always returns an empty list.**
  - `wasi:io`: output-stream `write`, `check-write`, `blocking-write-and-flush` and `blocking-flush`; input-stream
    `blocking-read`; `error.to-debug-string`.
  - `wasi:filesystem`: only `preopens.get-directories`, which returns **an empty list**, and descriptor drop.
  - **Absent**: clocks, random, sockets and http, and filesystem operations beyond the above.
- **WASI 0.3 (p3)**: the async component ABI is implemented and on by default, and `wasi:cli/run@0.3.0`-style
  exports can be called by name. A complete WASI 0.3 host is the v3 rung, not yet built.

## CLI quirks (observed)

- **`version` always prints `v0.0.1`** ([#1130](https://github.com/scttfrdmn/burroughs/issues/1130), open).
- **`inspect` does not understand components** (it reports them as malformed).
- **A component takes no `--dir`, `--scratch` or `--env`, and no function name**: each is refused by name as a usage
  error (exit 2), and its arguments go after `--`. A core module refuses `--features` by name (exit 2). These were
  [#1128](https://github.com/scttfrdmn/burroughs/issues/1128), fixed in `v0.12.0`, which also routes a component
  correctly whatever flags are given. Details are in the [CLI reference](cli.md#components-through-the-cli).
- **Caller mistakes on `Instance.Call`** (arity, type, unknown export) match no sentinel, and the CLI maps them to
  exit 1. Since `v0.12.0` a mistyped `Component.Call` argument is classed the same way, as the caller's error
  ([#1129](https://github.com/scttfrdmn/burroughs/issues/1129), fixed).

## Stale doc comments (observed)

These package doc comments say something the code no longer does
([#1130](https://github.com/scttfrdmn/burroughs/issues/1130), open). The behaviour described in these docs is the
code's:

- `ComponentConfig` says the value-carrying component surface "has not landed". It has.
- `ComponentKind` and `ComponentU32` say only `u32` crosses. `string`, `list` and `record` do too.
- `FeatureThreads` says `gate:threads`' default "is unchanged". It flipped on in `v0.10.0`.

## Platforms

- **Built and tested:** 64-bit Linux (`ubuntu-24.04`, x86-64 and arm64 runners) and macOS (development, plus a darwin
  CI job).
- **Windows** runs the test suite in a separate CI workflow that is measured, not a gate, and does not block a merge. Treat it as unsupported.
- **32-bit targets:** no claim.
- **Threads default**: ADR 0106's claim covers linux and darwin. `DefaultFeatures` itself does not check the OS, so a Windows build admits the grammar too, untested.

## Not started

- **Stack switching** (contract §7, the v2 rung).
- **A compiler.** No rung of the ladder adds one.
- **Production hardening against hostile modules** is a stated non-goal of contract v0.1 (§1.4).
