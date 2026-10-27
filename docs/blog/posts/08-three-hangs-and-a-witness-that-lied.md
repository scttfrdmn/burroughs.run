---
date: 2026-10-26
slug: three-hangs-and-a-witness-that-lied
categories: [Building Burroughs]
wip: true
---

# 8. Three hangs and a witness that lied

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

Concurrent code produces a particular kind of evidence: one red run in a few hundred, which doesn't come back when you re-run it. The cheap response is to push again and read the green as a refutation. Burroughs has a law against that (*"A re-run green doesn't refute a fail — explaining the fail does"*), and in the last week of October three investigations showed why.

<!-- more -->

## 1. The hang that wouldn't reproduce (#1106)

On 2026-10-25, during `make ci` on the branch right after the `gate:threads` flip, a component-model test hung once:

``` { .text .fragment }
--- FAIL: TestACancelRequestedWhileSuspendedIsDeliveredAtTheLoop (10.00s)
    cmthreads_cancel_test.go:30: cancel-suspended: run() did not return within 10s — the hang this witness exists to catch
```

The first attempts found nothing. **0 hangs in 453 attempts**: 350 isolated runs, 100 under parallel load, 3 replays with the failing shuffle seed. The issue body is careful: *"Zero in 453 says the window is narrow, not that it is closed."* It asked for two things: a goroutine dump at the bound, and a rate.

The chair's order was to **treat a one-off hang as a deadlock until shown otherwise**. The method was 10,000 iterations under `-race`, with a watchdog that dumps every goroutine when the bound expires. It hit **once in 10,000 on darwin/arm64** and not at all in 10,000 on janus x86-64. That one dump showed four parties all waiting on each other.

The bug was a classic **lost wakeup** in the component scheduler (`cmScheduler.drive`):

``` { .text .fragment }
 drive                         caller's fence              last writer
   │ pop fenced task's Thread      │                           │
   │──────────────────────────────▶│ refuse reader,            │
   │                               │ register a "kick"         │
   │                               │                           │ leaves
   │                               │ kick fires: ready queue   │
   │                               │ is EMPTY → sends nothing  │
   │ requeue Thread                │                           │
   │ (free() does not kick)        │                           │
   ▼                               ▼
 Thread sits queued forever; the parked task and its caller wait with it
```

The fix is one retry: after the requeue, try the reader once more. If the writer is gone by then, the Thread is admitted. If it's still there, its kick now finds a queue that holds the Thread. The witness holds the window open through a test seam on every run that reaches it, and with the retry removed it *"hangs, with the same dump."* Afterwards: 10,000 of 10,000 on each architecture.

The closing lesson: *"a refusal that registers its own wakeup must be re-checked after the thing it wakes is queued."* The `len(ready) == 0` check was correct as written. The ordering around it was not.

## 2. Impatience, or the memory model? (#896)

The second case looked similar. `TestStopBringsATailCallLoopToASafepoint` waits 30 seconds after `Resume` for a tail-call chain to finish, and it occasionally timed out on CI under `-race`. Issue #896 had been open since 2026-10-05, explicitly declining to just raise the bound: *"'the number was inconvenient' is the wrong moment to pick a new one."*

After #1106 the same method was applied: 10,000 `-race` iterations each of this test and a sibling, on both architectures, with a dump at the bound.

| Host | Tail-call test | Concurrent-Stop test |
|---|---|---|
| darwin/arm64, dev host | 9,998 pass, 2 timeouts | 9,994 pass, 6 timeouts |
| janus x86-64, idle | 10,000 pass, 0 | 10,000 pass, 0 |

All 8 dumps showed the same thing: the guest goroutine *runnable* inside the interpreter's hot path (`enterFrame`, `getLocal`, `binI32`, `branch` …), waiting for a CPU. That is impatience, not a lost wakeup.

Then **Scott asked a sharper question**: could darwin's timeouts come from ARM's weaker memory model? The stress couldn't say, because darwin (ARM) had run at about 0.67 processes per core and janus (x86) at about 0.5, so architecture and load were confounded. The answer was a 2×2 experiment, host by load, with 1,000 iterations per test per cell, each normalised by its own host's unloaded median:

| Arm | Tail-call median | p99 | Timeouts |
|---|---|---|---|
| darwin/arm64, 0.5 procs/core | 8.3s (×1.02) | ×1.06 | 0 |
| darwin/arm64, 0.67 procs/core | 8.7s (×1.07) | ×1.16 | 0 |
| janus x86-64, 0.5 procs/core | 17.1s (×1.46) | ×1.51 | 0 |
| janus x86-64, 0.69 procs/core | 21.2s (×1.82) | ×1.88 | 0 |

The finding: **contention, not the memory model.**

- The distribution moves with load and stays tight, with p99 within about 6% of the median in every cell. A memory-model defect would show up as a separate slow mode, such as a guest stuck on a stale value, and none appeared.
- At the stress's own load, darwin had **0 timeouts in 2,000**. All 8 original timeouts fell in roughly the first 10% of the stress, when other work was running on the dev host. A memory-model defect would not cluster in time.
- ARM wasn't even the sensitive side. Darwin's distribution was tighter than janus's in every cell.

The bound became **60s**, set from measurement as the chair ruled: the slowest passing post-`Resume` time measured was 22.1s (janus, 22 processes), giving 2.7× margin. The tests now log their post-Resume time and dump every goroutine at the bound, because *"a lost wakeup never finishes, so it still fails the bound."*

## 3. The witness that assumed its own premise (grave #1118)

The third case is a grave, and the most instructive of the three, because this time the engine was right and the test was wrong.

On 2026-10-26, main went red at `98dccb7` on the ubuntu-arm64 runner. The merge changed no engine lines, only scripts and a test.

``` { .text .fragment }
--- FAIL: TestATableFillHeldAcrossAParkSurvivesASiblingGrow
    oneload_test.go:193: table.grow returned 2097152 while another agent was part-way
    through a fill holding the old image; it must refuse to relocate (-1)
```

This was one of the "one-load witnesses" for ADR 0073 (post 5): `grow` must refuse to relocate while a sibling agent could hold the old image. The test arranged agent A to be mid-fill, then had agent B grow the table. It established "A is still filling" *by the order of two calls*: `Resume`, then B's grow.

That order doesn't guarantee anything. If A's fill finishes first, nothing holds the old image, B's grow correctly relocates, and the witness reports the very defect it exists to catch. The test's own harm check agreed that nothing was wrong: no null slot, and the fill completed into the live image.

The reproduction method was the one these investigations kept using: **widen the gap.** On the dev host, 0 of 200 isolated runs failed. A copy with a 200ms sleep between `Resume` and the grow failed **5 of 5**, with CI's exact message and a clean fill.

The repair makes the test *observe* its premise instead of arranging it:

- After B's grow returns, check whether A is still filling.
- A relocation with A **still filling** is the defect, definitively.
- A relocation after A returned is **inconclusive**, and the scenario reruns on a fresh instance.
- A floor fails the test if every scenario is inconclusive, so it can't pass by never deciding.

Neutered (with the engine's sibling check removed), it fails 10 of 10. With the gap widened, both arms end on the floor and never on the defect.

The closing comment states the lesson:

> **A witness must read its premise, not order it.** … two calls in sequence order nothing about a third agent's progress.

Its final detail is the most uncomfortable one: *"The file's own neuter note had already recorded the bad ordering in 2 of 10 runs, but under the neuter, so it read as a catch rather than as the flaw."* The evidence had been there from the start, recorded under a label that made it look like success.

## The common thread

The three cases had different causes: an engine bug, a test bound, a test premise. They were solved with the same tools:

1. **A rate, not a single re-run.** 10,000 iterations, not 453, and not one.
2. **A dump at the bound.** A hang that dumps its goroutines tells you whether anything is waiting, and on what.
3. **Widen the window.** A sleep in the gap turns a timing bug that 200 isolated runs never showed into 5 of 5.
4. **Separate the confounded variables before blaming the exotic one.** ARM's memory model was the interesting hypothesis, and contention was the true one. The 2×2 experiment was the cheapest way to tell them apart.
