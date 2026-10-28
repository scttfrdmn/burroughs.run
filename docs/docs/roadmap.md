---
wip: true
---

# Roadmap

***Subject to change**. This page states only what is decided or scheduled, without dates.
The authority is the GitHub tracker: milestones are the phase ladder, and an issue joins a milestone when it is
scheduled, not when it is filed. Read on 2026-10-28, after `v0.12.0`.*

## The phase ladder

The project moves through a sequence of artifacts. Each rung corresponds to sections of the contract:

| Rung | Contract | Artifact | State |
| --- | --- | --- | --- |
| **v0**: interpreter | — | decoder, internal form, validator, interpreter; the MVP core suite green; every Wasm 3.0 gate present with its default a recorded decision | **done**, at `v0.4.0` |
| **v1**: threads + safepoints | §§2–5 | OS-thread spawn, futex wait/notify, engine-native stop-the-world, the §4 boundary memory model and its litmus battery | **current phase.** Engine mechanism and battery built. Guest-started threads reach the public `wasip1` paths since `v0.12.0`; host-started threads are not built. |
| **v2**: stack switching | §7 | growable continuation stacks (a `morestack` analog), walkable at a safepoint for a guest GC | not started |
| **v3**: component model + WASI 0.3 | §6 | batch readiness (`poll_ready`), guest-pulled readiness, a host-owned event loop shared across component instances | the component model and async ABI are already on by default. The WASI 0.3 host is not built. |

Engine releases are `0.x` minors, each mapped to a milestone. `1.0.0` is reserved and is not on this page.

## Done in v0.12.0: guest-started threads, and four flips

`v0.12.0` (2026-10-28) delivered its milestone's headline, issue
[#1122](https://github.com/scttfrdmn/burroughs/issues/1122): *a fork-built Go program runs with real threads through
the public paths, with nothing supplied beyond the module*. It also fixed
[#1128](https://github.com/scttfrdmn/burroughs/issues/1128) and
[#1129](https://github.com/scttfrdmn/burroughs/issues/1129), and turned four Wasm 3.0 proposals on by default:
extended constant expressions, tail calls, multi-memory and memory64. Its open known limit on the threads side is the
Phase 4 fork's `LockOSThread` crash ([#1157](https://github.com/scttfrdmn/burroughs/issues/1157)).

The open v1 issues around it include the per-thread execution context (T-4's register-cost slot, #514),
engine-native safepoints and `stop(deadline)` as public surface (#515), and the contract's open question on
cancelling blocked host calls (§10.4, H-3, #13). Host-started threads (a spawn in the Go API) are not built.

## Default-off gate flips

**Two proposals stay gated off by default: GC and exception handling**, each with open work (#95; #172, #262). The
conformance harness runs a per-gate lane for each default-off gate, measuring what turning that gate on alone would
change, and the forecasts for the six original flips are on [#1131](https://github.com/scttfrdmn/burroughs/issues/1131).
Four of the six flipped in `v0.12.0`, each its own stamped event with a stated rollback (`BURROUGHS_<GATE>=0`), and
each listed in `CHANGELOG.md` under **Added** with its `gate:` name. **Of the 182 vectors that need two or more gates
at once, none was answered by those four alone**: each also needs EH or GC.

Turning a gate on by default and adding a public `Feature` to request it are separate things. No `Feature` is decided
for GC or exception handling.

## Later

- **WASI 0.3** (v3): the event loop and readiness clauses of contract §6. A milestone, v0.14.0 (proposed), puts it directly after `v0.12.0`, with no wasip2 step; its plan is not yet
  decided. The open component-model questions on this
  milestone include §10.2 (implement the canonical ABI in-house, or interoperate with wasm-tools/jco) and §10.6
  (`GOOS=wasip3` ABI negotiation).
- **Stack switching** (v2): contract §7, with §10.5 (the continuation-walk interface for GC scanning) open.
- **More component value kinds**: the component value surface grows when a guest needs a kind (ADR 0085's
  guest-driven rule). No specific next kind is scheduled.
- **Preview 1 retirement**: `WASIP1Config` is described as an on-ramp, to be retired once a Go toolchain that emits
  WASI 0.3 is usable (#688, now closed, was the recon of that condition).

## Not on the roadmap

- A compiler or JIT.
- Browser embedding (contract §1).
- Production hardening against hostile modules, for contract v0.1 (§1.4).
