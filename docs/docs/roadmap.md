---
wip: true
---

# Roadmap

***Subject to change**. This page states only what is decided or scheduled, without dates.
The authority is the GitHub tracker: milestones are the phase ladder, and an issue joins a milestone when it is
scheduled, not when it is filed. Read on 2026-10-26.*

## The phase ladder

The project moves through a sequence of artifacts. Each rung corresponds to sections of the contract:

| Rung | Contract | Artifact | State |
| --- | --- | --- | --- |
| **v0**: interpreter | — | decoder, internal form, validator, interpreter; the MVP core suite green; every Wasm 3.0 gate present with its default a recorded decision | **done**, at `v0.4.0` |
| **v1**: threads + safepoints | §§2–5 | OS-thread spawn, futex wait/notify, engine-native stop-the-world, the §4 boundary memory model and its litmus battery | **current phase.** Engine mechanism and battery built. Public exposure is the next step. |
| **v2**: stack switching | §7 | growable continuation stacks (a `morestack` analog), walkable at a safepoint for a guest GC | not started |
| **v3**: component model + WASI 0.3 | §6 | batch readiness (`poll_ready`), guest-pulled readiness, a host-owned event loop shared across component instances | the component model and async ABI are already on by default. The WASI 0.3 host is not built. |

Engine releases are `0.x` minors, each mapped to a milestone. `1.0.0` is reserved and is not on this page.

## Next: v0.12.0, guest-started threads

Milestone **v0.12.0 guest-started threads**, issue [#1122](https://github.com/scttfrdmn/burroughs/issues/1122):
*a fork-built Go program runs with real threads through the public paths, with nothing supplied beyond the module.*
This is the work that is meant to remove `v0.11.0`'s headline limit, that no public path can start a second thread
over a shared memory. Two defects these docs flag are on the same milestone:
[#1128](https://github.com/scttfrdmn/burroughs/issues/1128) (the CLI misroutes a component given arguments,
`--dir` or `--features`) and [#1129](https://github.com/scttfrdmn/burroughs/issues/1129) (a mistyped
`Component.Call` argument is classified `ErrUnsupported`). How the spawn door is exposed is a public-API decision, and it belongs to Scott Friedman.

The open v1 issues around it include the per-thread execution context (T-1 spawn on an OS thread, T-4's
register-cost slot, #514), engine-native safepoints and `stop(deadline)` as public surface (#515), and the contract's
open question on cancelling blocked host calls (§10.4, H-3, #13).

## Default-off gate flips

Six proposals are gated off by default: GC, exception handling, tail calls, memory64, multi-memory and
extended-const. Since just after `v0.11.0` (on main, not yet in a release; #1127) the conformance harness runs a per-gate lane for each one, measuring what turning
that gate on alone would change. Forecasts drawn from it have been posted for Scott's stamp on [#1131](https://github.com/scttfrdmn/burroughs/issues/1131), under the proposed milestone v0.13.0. **No flip has been
approved.** Each flip, if and when it happens, is its own stamped event with a stated rollback, and appears in
`CHANGELOG.md` under **Added** with its `gate:` name.

Turning a gate on by default and adding a public `Feature` to request it are separate things. Neither is decided for
these six.

## Later

- **WASI 0.3** (v3): the event loop and readiness clauses of contract §6. A milestone, v0.14.0 (proposed), puts it directly after v0.12.0, with no wasip2 step. The open component-model questions on this
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
