---
date: 2026-10-28
slug: guest-started-threads
categories: [Building Burroughs]
written: 2026-10-29
---

# 11. Guest-started threads

*Drafted 2026-10-29 from the `v0.12.0` tag, its `CHANGELOG.md` block, and the contract at that tag. Issue states are as read on that day and will move. Corrections welcome.*

Post 10 described `v0.12.0` as the next milestone, built around #1122 and Scott's goal: *"Go->Wasm->Burroughs runs a Go program in all its glory on an engine that can."* It shipped on 2026-10-28, the date of its `CHANGELOG.md` block (the tag itself is dated 2026-10-27 in Pacific time). The tag message names both halves: *"guest-started threads, and four Wasm 3.0 proposals on by default."*

<!-- more -->

## The door

A WASI preview-1 module that imports `"burroughs" "spawn"` and exports `mstart` now starts agents on OS threads over its shared memory, through `WASIP1Config.Run` and `burroughs run`. Each call to the import starts a new agent running `mstart(arg)`. Nothing is supplied beyond the module. A module that imports the door without exporting `mstart`, or with the wrong types, is refused at link, by name.

The contract names the door, in an append to §2 T-1: *"T-1's primitive is reachable from a guest."* The append says why it is an append: it *"names the guest-facing door to the primitive above and changes nothing T-1 already required."*

This lifts `v0.11.0`'s limit, *"no public path starts a second thread"*, for the WASIP1 path only. The `v0.12.0` block is exact about the rest: core `Instantiate` *"still supplies no imports (host-started threads are not built), and a component's `thread.spawn*` is still refused."*

## The witness

The end-to-end witness is a committed Go program built by the fork, run through the same function `burroughs run` calls. Its output must equal stock Go's, with a digest the test recomputes natively, and at least 3 agents must be sampled executing guest code at once. The block records 4. A `GOMAXPROCS = 1` build of the same source still matches the digest, and fails the floor.

The test, its numbers, its neuters and what it does not show are on [the v0.12.0 threads witness](../../docs/threads-witness.md) page, which this post does not restate.

Two things turned up while building it:

- **The first gauge could not see parallelism.** The peak number of *live* agents was the obvious measure, but the `GOMAXPROCS = 1` build still reached 3, *"because the Go runtime keeps Ms for its own work."* So the witness samples agents inside a call and not blocked instead.
- **The fork crashes on one pattern.** A guest that calls `runtime.LockOSThread` and then needs another M throws *"on a locked thread with no template thread"* (#1157). The throw is the fork runtime's own, and the witness does not pin its workers for that reason.

## Four flips

Post 9 ended with a per-gate board lane that gives what turning on *one* default-off gate does, and with its forecasts posted for Scott's stamp. Scott's answer, relayed in the chair's ruling of 2026-10-27, was *"Agree with your recommendation."* Four proposals were flipped, one stamped event and one ADR each, in this order:

| Flip | ADR | Rollback | Core board after it (pass / gated) |
|---|---|---|---|
| before | | | 60957 / 4187 |
| extended constants | 0107 | `BURROUGHS_EXTENDED_CONST=0` | 61027 / 4117 |
| tail calls | 0108 | `BURROUGHS_TAIL_CALL=0` | 61053 / 4091 |
| multi-memory | 0109 | `BURROUGHS_MULTI_MEMORY=0` | 61695 / 3446 |
| memory64 | 0110 | `BURROUGHS_MEMORY64=0` | 63769 / 1368 |

Each flip *"moved the core board by exactly its own lane's pre-registered reading, with 0 new failures, identically on darwin/arm64 and janus x86-64."* A control, `TestEachFlippedGateMovesOnlyItsOwnVectors`, compares the board with each gate on and off, file by file, and gained a row per flip.

**Exception handling and GC stay off**, with their open work (#95; #172, #262).

## A reading of the hardware

Post 6 quoted the `v0.11.0` block on the §4 battery's strength: *"Go's race detector and scheduling claims, not hardware weak-memory readings."* `v0.12.0` adds the reading (`scripts/hwlitmus.sh`, #1126). It runs the MP, SB, LB and IRIW litmus shapes at 2²⁰ trials per cell on darwin/arm64, janus x86-64 and Graviton3, over the engine and over a neuter build without ADR 0054's atomics.

- The engine shows **0** SC-forbidden outcomes everywhere.
- **Only Apple arm64 discriminates.** There the neuter shows SB's forbidden outcome in every run and MP's in most. Elsewhere, *"the interpreter's store-to-load latency hides the reorderings, and those zeros are not evidence."*

Contract B-MM-5 gains a non-normative note saying so. It changes no MUST.

One more finding came out of it: the pinned oracle, wasmtime 49.0.3, has no wasi-threads (removed upstream, and first absent from v46.0.0), so the wasmtime arm runs v36.0.19 LTS and the pin is unchanged.

## The rest of the release

- **Three contract appends**, on Scott's answer of 2026-10-28, *"no objection to all five, go ahead on the repo"*: T-1's door, an interpretation of R-2 (flagged as one), and B-MM-5's note.
- **ADR 0076 and ADR 0102 accepted**, on the same answer, and the decisions queue is no longer parked. Post 3 recorded both ADRs left `proposed` in `v0.11.0`. The `v0.12.0` block leaves that note as written, because it *"was true when it shipped."*
- **Three public-path fixes:** `burroughs run <component>` is routed whatever flags are given (#1128); a mistyped `Component.Call` argument is the caller's error, not `ErrUnsupported` (#1129); and a shared memory used as a Canonical ABI memory option is refused by name (#1141).

The board at the tag: **63769 pass, 0 fail, 0 unsupported, 1368 gated, 0 unimplemented** over 256 files. The all-gates-on lane reads 65109 pass, 0 fail, 0 gated. The threads lane is unchanged from `v0.11.0`, and its 41 failures are rows of a superseded corpus, *"with 0 engine defects."*

## What it did not do

The block lists six known limits, each with an issue: the fork's `LockOSThread` crash (#1157); `Close` with a zero bound returning `nil` on Windows (#945); the arm64 runtime fatal seen once (#1089); `array.new*`'s unpolled allocation, which applies only with GC on; a shared-everything-threads heap type refused as "malformed" rather than by name (#1120); and a Go guest's stop-the-world still depending on the fork's compiler-inserted loop checks (#1064, parked).

Three of the issues post 10 listed in the milestone did not ship in it. As read on 2026-10-29, #1132, #1133 and #406 are on the `v0.13.0 (proposed)` milestone.

Host-started threads are still not built, and a component's `thread.spawn*` is still refused.
