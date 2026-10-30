---
date: 2026-10-26
slug: whats-next
categories: [Building Burroughs]
written: 2026-10-29
---

# 10. What's next

*Drafted 2026-10-26 from the repository's records, and revised 2026-10-29 after `v0.12.0` shipped: its section is cut down to a pointer to post 11, and the sentences `v0.12.0` overtook are changed. Issue states are as read on the day each says and will move. Corrections welcome.*

## v0.12.0, shipped

When this post was drafted, the next milestone was `v0.12.0 guest-started threads`, built around #1122 and Scott's goal that *"Go->Wasm->Burroughs runs a Go program in all its glory on an engine that can."* It shipped on 2026-10-28, and post 11 covers what it delivered and what it did not.

<!-- more -->

Of the six open issues the milestone held as read on 2026-10-27 (UTC), #1122, #1128 and #1129 closed with the release. As read on 2026-10-29, #1132, #1133 and #406 are on the `v0.13.0 (proposed)` milestone.

## The rest of the ladder

```mermaid
flowchart LR
  v0["v0 interpreter<br/>✔ closed at v0.4.0"]:::done --> v1["v1 threads + safepoints<br/>in progress"]:::now
  v1 --> v2["v2 stack switching<br/>not started"]
  v2 --> v3["v3 component model + WASI 0.3<br/>much already landed, ahead of order"]:::ahead
  classDef done fill:#dfd,stroke:#393
  classDef now fill:#ffd,stroke:#996
  classDef ahead fill:#def,stroke:#369
```

The ladder no longer reads strictly left to right. The component model and WASI 0.3 async, nominally v3, landed in September and early October, behind gates that were then flipped, while v1 was still open. The component work numbers its own phases (Phase 2 components, Phase 3 async, Phase 4 the Go fork), separate from the v0–v3 ladder, and the `v0.6.0` notes state how they relate: *"phases 2–3 are the product, the fork is the thesis."* v2 (contract §7, growable continuation stacks) hasn't started. The roadmap reading of 2026-10-18 (#1046, a snapshot taken from the tracker and tree that *"replaces anyone's recollection of the roadmap"*) records that `binary.Features` *"has no stack-switching field"* and that the v2 milestone holds one issue, the §10.5 continuation-walk interface (#14).

## Contract clauses with no implementation yet

The same roadmap reading lists contract promises nobody has built. Each is tracked; the reading recorded #1047, #1048 and #1049 as parked. States below are as read on 2026-10-27 (UTC):

| Clause | Promise | State (issue) |
|---|---|---|
| §3 SP-3 | A deadline/timer wake facility, its channel disjoint from guest sync state (the clause the bell wrote) | no implementation (#1047, milestone `unscheduled`) |
| §6 R-1 | A batch readiness API, `poll_ready`, the netpoller analog | no occurrence in the tree (#1048, now on milestone `v0.14.0 (proposed) WASI 0.3`) |
| §8 M-2 | A decommit hook, so a guest's RSS isn't a ratchet | no implementation (#1049, milestone `unscheduled`) |
| §7 S-1–S-4 | Growable continuations, a native switch, a walkable stack set | not started (v2; #14) |
| §2 T-5.2 | A *host-side* join an embedder can call | exists only on the internal `interp.Instance` (#657) |
| §5 H-3 | A guest-visible cancel for blocked host calls | open (#13, v1 milestone); Scott's recorded lean is *"v1 minimal — shutdown-interruptibility is the MUST"* |

Two core proposals remain **default-off**: exception handling and GC. Six were off when this post was drafted. The other four (extended constants, tail calls, multi-memory and memory64) were flipped on in `v0.12.0` from the per-gate forecasts of post 9, which took the board from 4187 gated vectors to 1368 (post 11). The contract's M-4 says GC is implemented *"for conformance … and deliberately unoptimized. Go brings its own collector."*

## The distance to v1.0.0

`v1.0.0` is reserved for the v1 threads-and-safepoints milestone landing, and ADR 0004's amendment (approved by Scott on #795) sets four conditions on that landing. The v1 milestone itself still has four open issues as read on 2026-10-27 (UTC): #13, #514, #515 and #542.

1. **Every registered §4 litmus case landed or deferred with a live expiry.** Met at `v0.11.0`: 19 of 19, with no deferral left.
2. **The battery run on both a TSO and a weakly-ordered platform.** CI gates on linux x86-64, linux arm64 and darwin arm64.
3. **The memory-model reach stated at its measured strength** wherever the version's coverage is claimed. The `v0.10.0`/`v0.11.0` claim texts are written that way. Until `v0.12.0`, no public path started a second thread, so there was no public-threads claim to state. `v0.12.0` opened the WASIP1 path (post 11).
4. **The contract stable** (§1 non-goal 4). It is still v0.1, with amendments that each name what forced them.

A recon filed on 2026-10-26, #1126, examines *"a hardware weak-memory reading for the §4 battery"* and *"what it could establish,"* and raises a wasmtime core-threads option that the earlier #742 recon *"did not weigh."* It is explicit that this would not restore what #795 gave up: the wasmtime option could corroborate the battery's registrations for the guest→guest cases, not certify Burroughs, and *"Nothing here reopens #795."* The reading itself landed in `v0.12.0` as `scripts/hwlitmus.sh`, and of the three machines it ran on, only Apple arm64 discriminates (post 11).
## Open questions

Several things are unresolved, and the tracker says so:

- **Whether to report wasmtime's bug upstream** (#1045) is Scott's call, and the fixture is ready.
- **Windows** is compiled, run as a measurement, and outside every claim. Scott's 2026-10-10 ruling made it unsupported.
- **A Go runtime fatal seen once on CI's arm64 runner** (#1089) has not repeated, and is carried as a known limit rather than treated as resolved.

## A closing note on method

This series could be written only because the project's records are good enough to write history from. Nearly every sentence here traces to a commit message, an ADR's `Status:` line, a release block's "known limits," or an issue's closing comment. The same records explain the project's habit of correcting itself in public: the `v0.4.0` note that said *"it is not v1"* and was superseded, ADR 0002's falsified "free," the `v0.10.0` threads claim that `v0.11.0` corrected the next day.

The B5000 favored ALGOL by building the machine around it. Burroughs is trying to favour Go by building the *promises* around it first, and then checking, more carefully than most projects do, whether each promise is actually kept.
