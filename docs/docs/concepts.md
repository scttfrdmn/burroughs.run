---
wip: true
---

# Concepts

*Written against `v0.11.0`. The contract summaries below are simplifications. The normative text is
`docs/burroughs-contract-v0.1.md` in the repository, and wherever this page and that file differ, the file is
right.*

## The contract, in plain terms

Burroughs was designed around a document written **before** the engine: the *Burroughs Contract*, v0.1. It lists
what a host must promise a guest runtime like Go's. Most clauses trace back to a specific failure met while porting
the Go runtime to hosts that made no such promise.

The posture is in §0: **correctness-neutral, performance-partisan**. Any guest that conforms to the spec must run
correctly, and the upstream spec test suite is what checks that. Where a design tradeoff exists, the choice that
favours Go wins. "Partisan" applies only to API surface and optimization priorities, never to correctness.

The contract's non-goals (§1) are peak-throughput parity with optimizing compilers, browser embedding, ergonomics for
non-Go guests beyond conformance, and, for now, production hardening and sandboxing guarantees against hostile
modules. **Burroughs describes itself as a research engine first.** Keep that last non-goal in mind before running
untrusted code.

| § | Subject | What it promises (simplified) | State at v0.11.0 |
| --- | --- | --- | --- |
| §2 | Threads | Real OS threads for wasm threads (`spawn`). No main-thread special case, so any thread may block. Futex-backed `wait`/`notify`. A cheap per-thread slot (the Go `g` register analog). Defined exit, join and detach. | Built in the engine. **Not reachable from the public API**: no public spawn. |
| §3 | Safepoints | Preemption is the engine's job. `stop(deadline)` brings every guest thread to a safepoint in bounded time, without the guest instrumenting its own code. A thread blocked in a host call or a `wait` counts as stopped. | Built in the engine, **not public**. A Go guest's stop-the-world still depends on the fork compiler's loop checks (#1064). |
| §4 | Boundary memory model | Every host↔guest crossing is a full fence: what one side wrote is visible to the other after the crossing. | Litmus battery: 19 of 19 cases implemented, witnessed with Go's race detector and scheduling (not hardware weak-memory readings). |
| §5 | Blocking host calls | A blocking host call blocks only its own thread. Host calls never re-enter the guest on the caller's stack. | Engine-internal. The public API has no host-function registration. |
| §6 | Event loop (WASI 0.3) | Readiness is pulled by the guest's scheduler in batches (`poll_ready`), never pushed into guest code by callback. | Async component ABI on by default. The full WASI 0.3 host is the v3 rung. |
| §7 | Stacks | Stack switching with growable continuation stacks (a `morestack` analog), walkable by a guest GC. | **Not started** (v2). |
| §8 | Memory | `memory.grow` is cheap (reserve address space, commit on grow). A decommit hook. memory64 and multi-memory. WasmGC for conformance only. | Address-space reservation is in the code (ADR 0076, which is still marked `proposed`). memory64, multi-memory and GC are gated off. |
| §9 | Gates and conformance | Every proposal sits behind a gate. A gate defaults on only after its own spec suite is green on both a TSO (x86-64) and a weakly-ordered (arm64) platform. | See below. |
| §10 | Open questions | Settled by decision records, never silently in code. | Several still open. |

The contract versions **independently** of the engine. Every release states which contract it implements, and
`v0.11.0` implements contract v0.1.

## Gates

A **gate** is a switch around one proposal. Two rules govern gates (contract §9, the project's behaviour 4):

1. **Nothing defaults on without its own suite green.** A proposal lands behind a gate that is off.
2. **A flip is its own event.** Turning a gate on by default is a separate, stamped decision by Scott Friedman, with
   a forecast registered beforehand and a stated rollback. It never rides in the same change as the code it enables.

There are two kinds of gate: **decoder feature gates**, which decide which binary constructs are admitted (from
`internal/binary.DefaultFeatures`), and **config gates** for the component tiers.

| Gate | Kind | Default in v0.11.0 | Rollback / enable |
| --- | --- | --- | --- |
| SIMD (`0xfd`, `v128`) | feature | **on** | none |
| relaxed SIMD | feature | **on** | none |
| `gate:threads`: shared memory, `0xfe` atomics, wait/notify | feature | **on** (the claim covers linux and darwin; the code does not check the OS) | `BURROUGHS_THREADS=0`. `FeatureThreads` re-admits it on the wasip1 and component paths. |
| `gate:components` | config | **on** | `BURROUGHS_COMPONENTS=0` |
| `gate:async` (async component ABI) | config | **on** | `BURROUGHS_ASYNC=0` |
| `gate:cm-threads` (component-model threads, phase 1) | config | **on** | `BURROUGHS_CM_THREADS=0` |
| exception handling | feature | off | no public way to enable |
| GC | feature | off | no public way to enable |
| tail calls | feature | off | no public way to enable |
| memory64 | feature | off | no public way to enable |
| multi-memory | feature | off | no public way to enable |
| extended-const | feature | off | no public way to enable |

Wasm 2.0's features (reference types, bulk memory, multi-value, and the rest) are not gated.

A module using a default-off proposal is not broken, and Burroughs does not call it broken: it is refused as
`ErrGated` (CLI exit 6), naming the proposal. The repository's conformance harness runs an "all gates on" lane in
which no vector is left gated. Since just after `v0.11.0` (on main, not yet in a release) it also runs a per-gate lane for each default-off gate, which is
what each flip's forecast is registered against. The flips themselves wait on Scott's stamps.

**"Threads on" means the grammar, not threads.** Since the `gate:threads` flip (ADR 0106, released in `v0.10.0`),
shared memories and atomics are admitted. But no public path can start a second thread, so a public module's shared
memory has one agent. `v0.11.0`'s release notes correct `v0.10.0`'s claim on exactly this point.

## The vocabulary: trap, refused, declined, unsupported, gated

The project keeps these distinct because each one asks someone different to act.

| Word | What happened | Who acts | Go | CLI |
| --- | --- | --- | --- | --- |
| **trap** | The module was accepted and executed correctly, and the *program* went wrong: divide by zero, `unreachable`, out of bounds. | the program's author | `*Trap` (`Reason` is the spec text) | 4 |
| **refused**: malformed | The bytes are not a module this decoder can read. | whoever produced the bytes | `ErrMalformed` | 3 |
| **refused**: invalid | Well-formed, but it breaks a typing rule. The spec says reject it. | the module's author | `ErrInvalid` | 3 |
| **refused**: unlinkable | It imports something this path does not supply. | the embedder | `ErrUnlinkable` | 3 |
| **declined** | The validator has **no rule yet** for a construct, so the module was neither accepted nor refused. By default it runs anyway and `Instance.Decline()` names the construct. With `Strict` it is refused. | Burroughs (an incomplete validator) | `ErrDeclined` / `Decline()` | 3 under `--strict` |
| **unsupported** | The engine reached something it does not implement in this phase. On the core path this is at the point of use, not at load. (In `v0.11.0` a mistyped `Component.Call` argument also lands here, wrongly: [#1129](https://github.com/scttfrdmn/burroughs/issues/1129).) | Burroughs (an incomplete engine) | `ErrUnsupported` | 5 |
| **gated** | Well-formed, and the spec would accept it, but this build has that proposal switched off. | whoever chooses the gates | `ErrGated` | 6 |

Why **declined** is its own outcome: the validator has been arriving in slices. Calling an unchecked module
"invalid" would refuse modules the engine runs correctly, and calling it "valid" would claim a check that never
happened. Decision 0029 records why running a partly unvalidated module is defensible in a Go host. Today no module
in the default feature set is known to trigger a decline, so you are unlikely to see one.

Component calls add four more outcomes: **cancelled** (`ErrCancelled`, 7), **closed** (`ErrComponentClosed`, 8),
**close incomplete** (`ErrCloseIncomplete`, 9) and **locked** (`ErrInstanceLocked`, 4, because a trap caused it).

## Conformance: the spec suite board

The upstream WebAssembly spec test suite is the oracle. The project deliberately keeps spec-suite figures out of its
README, because a number typed into prose goes stale. Ask the instrument instead, from a checkout:

``` { .console .fragment }
$ make spec-tests                                    # vendor the suite (gitignored)
$ go test ./internal/spec/ -run TestPhase1Files -v   # print the board
$ make conformance                                   # assert it, default lane and all-gates-on lane
```

*These commands were not run for this page.* For reference, the `v0.11.0` release notes record the core board as
**60957 pass, 0 fail, 0 unsupported, 4187 gated, 0 unimplemented over 256 files**. The separate threads lane reads
576 pass and 41 fail, and all 41 failures are rows from a superseded corpus, with no engine defects. Those are that
release's figures, quoted from `CHANGELOG.md`. Run the board for current ones.

The board has five verdict columns:

- **pass**: answered, and right.
- **fail**: answered, and wrong. This is the work plan.
- **gated**: the engine could answer, but the proposal's gate is off.
- **unimplemented**: the harness asked, and no engine component exists to answer.
- **unsupported**: the harness cannot yet *ask*. This column is at zero, which is a fact about the harness's
  vocabulary, not an engine milestone.

Component-model behaviour is checked against two reference oracles: the component model's reference implementation
and a pinned wasmtime (49.0.3, checksummed since `v0.11.0`).

## Decisions and where they live

Design choices are recorded as ADRs in `docs/decisions/NNNN-*.md` **before** they are implemented. An ADR's
`Status:` cites the approval that accepted it, and stays `proposed` while there is no stamp to cite (0076, 0082 and 0083 still read `proposed`). The numbers these docs cite (0029 for the public boundary, 0082 for
refuse-at-link, 0083 for capabilities, 0085/0097 for the component value surface, 0088 for `Feature`, 0091 for
scratch grants, 0101 for one-task-at-a-time, 0106 for the threads flip) all resolve there. Work, bugs ("graves") and
roadmap live in GitHub issues and milestones.
