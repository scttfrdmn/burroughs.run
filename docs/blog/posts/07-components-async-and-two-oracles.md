---
date: 2026-10-21
slug: components-async-and-two-oracles
categories: [Building Burroughs]
wip: true
---

# 7. Components, async and two oracles

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

## 60,957 passing assertions and no program

ADR 0080 (2026-09-07) begins with an uncomfortable fact:

<!-- more -->

> Burroughs has 60,957 passing spec assertions and has never executed a program anyone wrote outside the project's own fixtures. The corpus proves compliance; it does not prove the engine is *usable*.

The obvious first program was Go itself. Stock Go has `GOOS=wasip1 GOARCH=wasm`. WASI preview 1 is an older interface than the WASI 0.3 the contract targets, so the ADR framed this carefully as *"a usability spike,"* not the start of v3.

As usual, the binary disagreed with the estimate, in two ways:

1. A Go hello-world declares **15** distinct preview-1 imports, not the roughly 10 its behaviour suggests. The extras are Go runtime startup probes (`sched_yield`, `poll_oneoff`, `random_get` for hashmap seeding before `main`, and so on). The linker refuses any unsatisfied import, so all 15 had to be supplied, even as stubs.
2. **It uses no Wasm atomics or shared memory.** A first regex counted 1132 `atomic.` hits, but they were `runtime/internal/atomic.*` *symbol names*. The ADR's lesson: *measure with the instrument, not a regex.* Stock Go's wasip1 port runs goroutines cooperatively on one Wasm thread.

The same day, a commit recorded the result: *"Burroughs executes a third-party-compiled program for the first time: a Go GOOS=wasip1 guest reaches main, writes stdout via fd_write, exits 0."* Within the month, `burroughs run` executed wasip1 commands from the CLI, a read-only filesystem with capability preopens landed (ADR 0083), and a confined writable scratch directory followed (ADR 0091, which leaves confinement to Go's `os.Root` instead of hand-written path resolution).

## Into the component model

The component model is the WebAssembly packaging layer above core modules. Components declare typed interfaces in WIT, and the **Canonical ABI** defines how strings, lists and records are *lifted* out of one module's linear memory and *lowered* into another's. WASI 0.3, ratified in June 2026 according to the contract, builds async into this layer, with `future<T>`, `stream<T>` and a host-owned event loop. That is v3 on the phase ladder.

The tiers opened in quick succession, each with its own gate and ADR:

| Gate | ADR | Opened | Default on | What it covers |
|---|---|---|---|---|
| `gate:components` | 0084 | 2026-09-09 | 2026-09-11 (Scott's stamp on #720) | Component loader; Canonical ABI lift/lower for value types, sync only |
| `gate:async` | 0086 | 2026-09-11 | 2026-09-18 (#792) | The async Canonical ABI: subtasks, waitable sets, streams, futures, cancellation |
| `gate:cm-threads` | 0100 | 2026-10-11 | 2026-10-17 (Scott's stamp on #999) | Phase 1 of shared-everything threads: eight built-ins, `0x26`–`0x2d` |

Each ADR pins the spec at a fixed commit (`WebAssembly/component-model` at `2bed77e`, which has no tags) because, as ADR 0084 puts it, *"every lift/lower decision below rests on which text was read."*

The milestones read like a hello-world sequence, which is roughly what they were:

- **A component prints "Hello, world!"** (`p3hello`), *"byte-identical to wasmtime 48.0.1."*
- **`p3async-hello`**, a Rust `wasm32-wasip3` guest, *"instantiates, runs, and writes `Hello, world!\n` — checked byte-for-byte against the committed wasmtime reading."* It was driven there one async built-in at a time, with each CHANGELOG increment ending *"`p3async-hello` now advances to …"*.

`v0.6.0` (2026-09-18) summed it up: *"Burroughs runs programs compiled by a third-party toolchain against WASI 0.3 — components and the async Canonical ABI, both gates on by default, two guests, byte-identical to an independent engine's committed reading."*

## A choice deferred to a guest

ADR 0086 refused to choose between the two async-lift ABIs. In **stackless callback**, the guest returns to the host at every suspension and is re-entered through a callback. In **stackful**, the engine keeps a suspended guest stack. As ADR 0092 later put it, *"Picking wrong is not a refactor,"* so the choice was deferred *"to the first guest that actually lifts an async export."*

When that guest existed, built with `wit-bindgen` 0.62.0, ADR 0092 (2026-10-02) read the answer off its bytes: **stackless callback**, *"because the first guest that lifts async chose it."* The contract was updated to match. S-1 (growable continuation stacks) stays unconsumed *"until a guest async-lifts an export stackfully."* That note was itself corrected once, because its first wording expired on the wrong event. The record says it was *"found by building that guest, not by re-reading the note."*

## Two oracles that disagree

The component tier has *two* candidate oracles, which is both a gift and a trap:

- **The pinned reference**, `definitions.py`: the component-model spec's executable Python model of the Canonical ABI.
- **wasmtime**, the major independent engine, pinned to a specific release.

They don't always agree. ADR 0102 (2026-10-12, an interim rule from the chair, still `proposed` with no stamp on record) records two cases where Burroughs had sided with *different* oracles without saying why. In one, the toolchains were stricter than the pin, and Burroughs refused. In the other, the pin was more permissive than wasmtime, and Burroughs admitted. The interim rule:

> The pinned reference defines the semantics; a toolchain's stricter refusal is followed unless Burroughs implements and witnesses what the pin admits.

The pin isn't easy to use as an oracle, either. Its scheduler picks threads with `random.choice`, and a trap raised on a thread's own Python thread leaves the thread that resumed it blocked forever. So `refread.py` runs every scenario under several seeds and profiles, reports outcomes as *sets*, records `HANG` after a bound, and is calibrated first on a fixture wasmtime *can* read: *"A scenario that has not been calibrated is not an oracle."*

## A bug in the other engine

Building `gate:cm-threads` turned up a wasmtime bug. A thread created with `thread.new-indirect` that blocks in a synchronous `waitable-set.wait` on a subtask event crashes wasmtime with:

``` { .text .fragment }
BUG: delivering non-present event
location: crates/wasmtime/src/runtime/component/concurrent.rs:948
```

Issue #1045 has the reproduction (`wasmtime-bug-repro.wat`, expected `5`), the narrowed cause and a diagnosed patch. It reproduces on 49.0.2 and 49.0.3, and on 2026-10-26 it was confirmed **still present in 50.0.0** (now at `concurrent.rs:984`). A source-built 49.0.3 *with* the patch reads `1000`, matching the pin's reading on the SP-5/H-4 fixture. Per the chair's ruling, as #1045 records it, that reading is *"corroboration only. The pin is the oracle."*

It has **not been filed upstream**. Filing is an outward act under Scott's name, so it is Scott's decision, and #1045 is open for exactly that. Meanwhile `build.sh` asserts that stock wasmtime *still* hits the bug, so a fixed release will fail the script and force a re-read.

That incident led to a smaller change for `v0.11.0`. The installed wasmtime had moved under the committed readings twice, once through a patch release and once through Homebrew's upgrade to 50.0.0. So the oracle is now fetched from its official release and checked against a committed SHA-256 (`scripts/oracle-fetch.sh`), *"so a package-manager upgrade can no longer move it under the readings."*

## Correct behaviour as a default change

The component releases are unusual in that most of their headline items are **default-behaviour changes made to match the reference**, and each says so. From `v0.8.0` and `v0.9.0`:

- Concurrent `Component.Call`s on one component **execute one at a time** (ADR 0101), because the reference runs one thread at a time and interleaves only at suspension points. This began as a single `-race` report on Windows that darwin didn't reproduce in 8,200 runs.
- **A trap locks the component instance**, and every later call returns `ErrInstanceLocked` wrapping the trap (ADR 0105), *"because the pin locks the store."*
- A callee that returns a value after cancellation was requested now resolves `RETURNED` with the value, instead of losing it.

Each says that no gate restores the old behaviour, because it was a conformance defect, not a feature.

## The fork

One more artifact sits beside the engine: **a fork of the Go toolchain** (ADR 0087, stamped 2026-09-17). The component track numbers its own phases (Phase 2 for components, Phase 3 for async), separate from the v0–v3 ladder, and the fork is Phase 4. It compiles Go to Wasm with real threads, binding `newosproc` to Burroughs' spawn under `GOEXPERIMENT=burroughsspawn`. The `v0.6.0` release line is *"phases 2–3 are the product, the fork is the thesis."*

The fork's first guests corrected an assumption almost at once. ADR 0088 records that a fork-emitted Go component, which wasmtime ran, was refused by Burroughs, because *"Go's compiler emits atomics unconditionally — 1148 operations in a hello-world that starts no goroutines."* In the engine, atomics and threads were one gate. In Go's runtime they are not. (The stock `wasip1` build from earlier in this post had none, which is why that guest never needed the gate.)

The charter is frank about the fork's evidence: the threaded tier has **no independent engine** to check against, because the shipped Canonical ABI has no `thread.spawn`. Its only memory-model oracle is the §4 battery. ADR 0104's later finding, that the fork's GC stops the world through its own compiler-inserted checks and never calls the engine's `Stop`, is recorded in post 5. Post 9 covers what happens when the fork's guests meet the public API.
