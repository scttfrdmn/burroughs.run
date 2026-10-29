---
wip: true
---

# Why Burroughs?

*Written against `v0.12.0`.*

A Go program compiled to WebAssembly carries a runtime, and that runtime expects things of its host that a
WebAssembly host is not generally designed to give. Each row below is one of those mismatches. The second column is
what the Go runtime expects. The third is what Burroughs' [host contract](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md)
commits to. The last column is the evidence: a test, a decision record or a contract clause. Where Burroughs' side is
still future work, the row says so and links the milestone that owns it.

| Mismatch | A generic Wasm host can assume | The Go runtime expects | Burroughs' choice | Evidence |
|---|---|---|---|---|
| **Blocking OS threads** | Threads are host-managed workers that exchange messages, and a blocking call can hold up an event loop shared by everything. | It creates OS threads on demand (`newosproc`), and a goroutine's blocking call blocks only its own thread. | A guest-callable spawn, each new agent backed 1:1 by an OS thread over the shared memory (T-1). A blocking host call blocks its thread only (H-1). **Built**, on the WASI preview-1 path. | [§2 T-1](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md#2-threads), [§5 H-1](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md#5-blocking-host-calls); [ADR 0068](https://github.com/scttfrdmn/burroughs/blob/main/docs/decisions/0068-spawn-drops-0056s-walk-and-refuses-the-two-cases-a-per-instance-world-cannot-express-because-a-thread-belongs-to-exactly-one-stop.md); [the threads witness](threads-witness.md) |
| **Stop-the-world and safepoints** | Core Wasm defines no way to stop a running agent, so a guest that must be stoppable instruments its own code. | It stops every goroutine at a safepoint to collect garbage, and a thread inside a system call already counts as stopped. | Engine-native safepoint polls at loop back-edges and calls, so a guest need not instrument itself (SP-1). A thread blocked in a host call or a `wait` counts as at a safepoint (SP-2). **Built in the engine, not public API.** A Go guest's own collections still rely on the fork compiler's inserted loop checks ([#1064](https://github.com/scttfrdmn/burroughs/issues/1064), parked). | [§3](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md#3-safepoints-and-preemption); [ADR 0074](https://github.com/scttfrdmn/burroughs/blob/main/docs/decisions/0074-stop-waits-on-sp-1s-own-predicate-over-the-caller-marks-because-an-arrival-is-a-caller-and-the-protocol-named-neither-end-of-it.md); [`TestStopBringsAGuestLoopToASafepointAndResumeLetsItFinish`](https://github.com/scttfrdmn/burroughs/blob/main/internal/interp/safepoint_test.go) |
| **Boundary memory ordering** | On the browser host, a notify orders only the notified word, so another field's store can lag behind the woken agent (the contract's own §4 note). | A woken goroutine sees everything written before the wake, with no spin to wait it out. | Every host↔guest transition is an acquire or release edge over the whole shared address space: a sequentially consistent fence at the boundary, both directions (B-MM-1). **Built.** Its litmus battery has 19 of 19 cases implemented. | [§4](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md#4-the-boundary-memory-model); [ADR 0106](https://github.com/scttfrdmn/burroughs/blob/main/docs/decisions/0106-gate-threads-defaults-on-with-burroughs-threads-0-as-the-rollback-because-the-4-battery-has-no-hole-and-the-off-state-stays-witnessed.md); [#10](https://github.com/scttfrdmn/burroughs/issues/10); [`TestAResumedAgentSeesASiblingFieldWrittenBeforeTheNotify`](https://github.com/scttfrdmn/burroughs/blob/main/internal/interp/battery_test.go) |
| **Runtime-managed stacks** | A continuation's stack is fixed at creation, which forces a goroutine model to overallocate or die (the contract's own §7 note). | Every goroutine starts with a small stack that grows on demand (`morestack`). | Growable continuation stacks, created small and grown through an engine hook, the `morestack` analog (S-1). **Future work:** stack switching has not started. | [§7 S-1](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md#7-stacks); milestone [v2 stack switching](https://github.com/scttfrdmn/burroughs/milestone/4) |
| **Netpoller readiness** | Readiness is delivered by callback into guest code, as an event handler. | The scheduler pulls ready work in batches from its netpoller when it looks for something to run (`findRunnable`). | Batch readiness, `poll_ready(buf, max, deadline)`, pulled by the guest scheduler and never pushed by callback (R-1, R-2), on WASI 0.3's host-owned event loop. **Future work:** `poll_ready` is not built. | [§6](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md#6-the-event-loop-and-readiness-wasip3); milestone [v3 component model + WASI 0.3](https://github.com/scttfrdmn/burroughs/milestone/5) |

## How the claims are tested

Three things together make the experimental apparatus. Burroughs is one. The **Phase 4 Go fork** is a modified Go
toolchain whose `wasip1` port starts real threads through Burroughs' spawn import; it is not published. The third is
the **Go-runtime torture suite** ([#406](https://github.com/scttfrdmn/burroughs/issues/406)): Go programs built with
the fork, each of whose observable results must match the same program built with stock Go.

None of these rows is a performance claim. Burroughs is an interpreter, and the contract's posture is
correctness-neutral and performance-partisan: any spec-conforming guest runs correctly, and where a design tradeoff
exists, Go wins.
