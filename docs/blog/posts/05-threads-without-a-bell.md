---
date: 2026-09-07
slug: threads-without-a-bell
categories: [Building Burroughs]
written: 2026-10-26
---

# 5. Threads without a bell

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

## The phase changes, quietly

`v0.4.0` (2026-08-28) closed v0, and its release note said firmly: *"it is not v1. … v0 closing means v0's conditions are discharged, not that the next phase has begun."* That was true when written: none of the §§2–5 threads work had started. Two days later it had.

<!-- more -->

Saying which phase the project was in turned out to be a governance problem of its own. Scott ruled on the #534 review that *no value naming v0 is true* any more. But that ruling lived only in the agent's brief, and the one public document that spoke about the phase, the `v0.4.0` note, denied it. ADR 0099 (2026-10-08) was written to fix that. Its title says what it does: *"the project is in the threads phase, recorded outside the agent brief because a phase claim needs a citable home."* The old release note was kept word for word, with a superseded marker beneath it.

## What v1 has to deliver

v1 is contract §§2–5: real threads, a stop-the-world, a boundary memory model, and blocking host calls that block only their own thread. For Go this is the core of the thesis: goroutines spread over many OS threads, a GC that stops the world, and syscalls that park an M while the others keep running.

The work came in a fast run of ADRs, twenty-seven of them (0050 to 0076) from 2026-08-30 to 2026-09-06. A selection:

| ADR | Date | What it decided |
|---|---|---|
| 0050 | 08-30 | A per-thread context object, reached by one pointer on the stack |
| 0051 | 08-31 | The 67 atomic instructions become sequentially consistent word operations |
| 0054 | 08-31 | *Every* aligned guest access becomes atomic |
| 0058 | 09-01 | The memory image is published through an atomic pointer |
| 0059 | 09-02 | The safepoint poll, guarded at the pc assignment |
| 0060 | 09-02 | The futex queue hangs off the memory, keyed by effective address |
| 0061 | 09-02 | `grow` serialises on its own mutex |
| 0064 | 09-04 | The bulk and SIMD region stays plain, confined by a checked enumeration |
| 0067 | 09-04 | A caller count joins the blocked mark |
| 0068 | 09-04 | `Spawn`: one OS thread per Wasm thread |
| 0069 | 09-05 | Host functions, and `Instance.Close` |
| 0071 | 09-05 | Thread exit, join and detach (T-5) |
| 0073 | 09-05 | `grow` refuses to relocate when a sibling could hold the old image |
| 0074 | 09-06 | `Stop` waits on SP-1's predicate over the caller marks |
| 0076 | 09-06 | Memory reserves address space with an anonymous mapping (§8 M-1) |

`v0.5.0` (2026-09-07) shipped all of this **behind `gate:threads`, off by default**.

## The guest's races must not be the host's

The first real fight was about data races. In a Wasm threads program, two guest threads may race on shared memory; the spec allows that and says what may be observed. But Burroughs is written in Go, and if a guest race becomes a Go data race inside the engine, Go's memory model makes no promises at all, and the race detector reports noise on every run.

Scott's ruling on the #566 review: *"the principle is that the guest's data races must not be the host's."* The cost was real. ADR 0054 made every aligned guest load and store a `sync/atomic` operation, and its consequences section gives the headline: *"amd64 aligned stores pay +13.72% uncontended … arm64 pays nothing measurable on any row."* (The landed mechanism later measured +10.16% on that row.)

The decision went back and forth. Scott had first declined a cheaper version because only a *test* property seemed to be at stake (a clean race detector). The agent then found that no sound way existed to apply atomics only where memory is actually shared. Scott's reply changed the grounds:

> With no sound gate available the racy region can't be confined at all, so the alternative isn't cheaper plain accesses — it's a permanently unbounded racy region that #10 could never certify anything against. That's foundational, not testability.

The principle was later given an exception, with conditions. Bulk operations like `memory.fill` and the SIMD loads stay plain (ADR 0064), because the threads proposal itself allows a racing fill to be seen torn. The exception holds only *"where the guest model itself permits the tear and the region is statically enumerable with the enumeration asserted by a control."* Where possible, the defect was made impossible to write rather than checked for. ADR 0065 moved slice headers *inside* the published image structs so the old racy assignments became compile errors. The law that came from it: *"An enumeration held by construction outranks one held by a scan."*

## Growing memory under a running thread

`memory.grow` is where threads hurt most. If growth allocates a bigger array and copies into it, a thread still running on the old array writes into memory nobody will ever read again. Issue #586's title states it: *"Relocation is memory-safe and not coherent: a thread left on an abandoned image."*

The fix came in two parts:

1. **Don't relocate under a sibling.** ADR 0073: `grow` refuses to move a memory while another agent could hold the old image (ADR 0075 does the same for tables), returning `-1`, the spec's legal "growth failed" answer. The boundary accessors take the growth lock.
2. **Mostly, don't relocate at all.** ADR 0076 implemented contract §8 M-1, which requires growth to cost O(pages touched). A memory now reserves its address space with an anonymous `mmap` and the kernel commits pages as they're touched. The Go allocator became *"the fallback rather than the mechanism."* Scott, on unscheduling other work, called this *"the dissolving move at platform level."*

## Futexes

T-3 asks for `memory.atomic.wait`/`notify` with OS-level wake latency. ADR 0060 hangs the futex queue off the memory itself, keyed by effective address. A pointer key was rejected because it *"would borrow its soundness from another package."* B-MM-2, the clause written in response to D20, requires that a wake publishes every earlier write by the waker, not only the futex word.

## Stopping the world

Here is the mechanism Go's GC needs, as built inside the engine: `Instance.Stop(deadline)` and `Resume`. No public API exposes it yet.

```mermaid
sequenceDiagram
    participant H as Host (calls Stop)
    participant A as Agent A (running guest code)
    participant B as Agent B (blocked in a host call)
    participant C as Agent C (parked in atomic.wait)
    H->>A: set stopReq
    H->>B: set stopReq
    H->>C: set stopReq
    Note over B,C: blocked/parked already count<br/>as at a safepoint (SP-2), and are not woken (SP-4)
    A->>A: next back-edge or call entry: poll sees stopReq
    A-->>H: parks at the safepoint
    Note over H: predicate holds: parked + blocked ≥ callers, per thread
    H->>H: Stop returns; host may mutate memory
    H->>A: Resume (an acquire edge, SP-6)
    Note over B: B returning from its host call re-enters<br/>through a boundary that observes the stop (SP-2)
```

Some details:

- **The poll is one atomic load**, reached from every back-edge (the fourteen `jumpTo` arms in the interpreter loop, kept complete by a test asserting that every pc assignment goes through `jumpTo`) and every call entry. ADR 0059's measured cost was a geomean of **+0.38% on arm64 and +0.73% on amd64**. ADR 0059's phrase for the design: *"straight-line code pays nothing at all."*
- **The predicate is about callers, not threads.** ADR 0067 found that SP-2's subject is a thread's *callers*: a thread re-entered from the host has more than one. `Stop` returns once, per thread, `parked + blocked >= callers` (ADR 0074). On deadline it returns `ErrStopDeadline`, naming each thread still running.
- **SP-6 was added on 2026-09-12** (Scott's stamp on #743) after the battery recon found that B-MM-1 said nothing about the `Resume`-after-`Stop` crossing. A GC that moves objects during a stop needs the resumed mutators to see everything, through ordinary loads.

This is also where an honest weakness turned up later. ADR 0104 (2026-10-18) re-read SP-1 against the Go toolchain fork (post 7) and found that *"the fork never calls the engine's `Stop`"*. Its GC uses the compiler's own inserted loop checks, *"exactly what SP-1's last sentence forbids: the guest self-instruments its code generation to be stoppable."* That half (SP-1 "reading (b)", #1064) is parked. Reading (a) was built for `v0.9.0`: bulk memory, table and GC-array instructions now poll as they go, so a big `memory.fill` doesn't delay a `Stop`. One named limit remains: `array.new*`'s single backing allocation, which takes more than 1ms at about 64 Ki elements on janus.

## Blocking, joining, closing

The rest of the §2/§5 surface came together in early September: host functions that mark their thread blocked for the duration of the call (ADR 0069), so a blocking host call stops only its own thread (H-1). Then T-5's exit, join and detach (ADR 0071): `Instance.Join`, `Instance.Fault`, and a `Close` that ends every thread and waits. Contract T-5.1 says threads are detached by default, and an unobserved exit *"MUST leak neither a wasm thread nor an OS thread."*

## Where v0.5.0 left it

The `v0.5.0` block is careful about what had *not* been shown. `gate:threads` was off. The §§2–5 litmus battery had *"a landed test for 5 of its 11 registered cases (T-1, T-2, T-3, and the two B-MM-2 cases), with the other six blocked."* And by Scott's #670 ruling, the rest of the battery was unscheduled.

The battery is the subject of the next post. It is the most unusual instrument in the project, because it is graded against an oracle the project wrote itself.
