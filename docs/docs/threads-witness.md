---
wip: true
---

# The v0.12.0 threads witness

*Written against `v0.12.0`. Every figure here comes from the witness test,
[`internal/wasi/threads012_test.go`](https://github.com/scttfrdmn/burroughs/blob/v0.12.0/internal/wasi/threads012_test.go),
from its guest's [`PROVENANCE`](https://github.com/scttfrdmn/burroughs/blob/v0.12.0/internal/wasi/testdata/threads012/PROVENANCE),
or from the `v0.12.0` block of the [changelog](https://github.com/scttfrdmn/burroughs/blob/v0.12.0/CHANGELOG.md).*

**The claim:** in `v0.12.0`, a fork-built Go program runs through `burroughs run` with `GOMAXPROCS > 1`, several
agents executing guest code at once on OS threads, and garbage collection across them.

The test is `TestAForkBuiltGoProgramRunsWithRealThreadsThroughThePublicPath`. It runs a committed Go program through
`Run`, the function behind both `WASIP1Config.Run` and `burroughs run` on a command, with nothing supplied beyond the
module.

## What the program computes

The guest's source is committed beside it
([`guest.go.txt`](https://github.com/scttfrdmn/burroughs/blob/v0.12.0/internal/wasi/testdata/threads012/guest.go.txt)).

- It sets `GOMAXPROCS(4)` and starts **4 workers**. Each runs a **200,000-round** hash chain from its own seed, and
  main folds the four results into one 64-bit digest.
- Alongside them it builds a ring of **50 nodes**, each with a finalizer and a checksum over its own fields and its
  successor's id. A goroutine rewrites nodes **600** times while main forces **8 collections**. Then main walks the
  ring and checks every checksum.

Its stdout must be exactly:

``` { .text .fragment }
digest 194b925792819171
workers 4
gc VERIFIED nodes 50 finalized 0 collections 8
```

- **The digest is recomputed with stock Go inside the test**, from the same chain function, so a wrong answer cannot
  agree with itself.
- **The GC line is fixed by construction:** every checksum verified, every node visited, and no finalizer run on a
  node that is still reachable.

## How "at once" is measured

While the program runs, the test samples the engine's count of agents executing guest code **every 200µs** and keeps
the largest value. At least **3** must be seen at once.

The floor was set from a measurement on darwin/arm64 (2026-10-27):

| Build | Most agents executing at once | Runs |
|---|---|---|
| the witness, `GOMAXPROCS(4)` | 4 | 4 of 4 |
| the same source with `GOMAXPROCS(1)` | 2 | 4 of 4 |

So 3 sits above the single-P build and below the witness. The peak count of **live** agents is logged too, but it
could not discriminate: the `GOMAXPROCS(1)` build still reached 3 live agents, because a Go runtime keeps OS threads
for its own work.

## The neuters, each watched failing

- **The spawn door not served:** the module fails at link, with *unknown import: "burroughs" "spawn"*.
- **`GOMAXPROCS(1)`:** the digest still matches, and the floor fails, with *at most 2 agents executed guest code at
  once, below 3*. So the floor measures parallelism, not the answer.
- **The garbage-collection half** has its own neuter in `TestClause3GCAcrossTwoAgents`: the fork built with
  stop-the-world disabled.

## What it does not show

- **"Executing at once" is a sample, and it counts the Go runtime's own agents.** The gauge counts agents running
  any guest code, and a Go guest's runtime (its scheduler and collector) is guest code. So the floor shows several
  OS threads running the program at once, and that the single-P build has fewer. It does **not** show that the four
  hash chains themselves overlapped. A later suite found the same gauge reading 4 for a program whose workers return
  at once (#1172).
- **It needs the Phase 4 Go fork.** The guest is built with a modified Go toolchain that starts its threads through
  Burroughs' spawn import (`GOEXPERIMENT=burroughsspawn`). Stock Go's `wasip1` port is single-threaded. The fork is
  part of the experimental apparatus, with Burroughs and the [Go-runtime torture suite](https://github.com/scttfrdmn/burroughs/issues/406),
  and it is not published.
- **A guest that calls `runtime.LockOSThread` and then needs another thread crashes** on the fork, with *on a
  locked thread with no template thread*
  ([#1157](https://github.com/scttfrdmn/burroughs/issues/1157)). The witness doesn't pin its workers, for that
  reason.
- **`burroughs run` itself is not under test here.** The test calls the function `burroughs run` calls. Running the
  committed guest through the `v0.12.0` CLI by hand prints the same three lines and exits 0. That is an observation,
  not a test, and it samples no concurrency.
- **One platform set the floor.** The floor's measurement was taken on darwin/arm64. CI runs the witness on linux
  amd64, linux arm64 and macOS, and on the non-gating Windows lane. The `v0.12.0` claim names linux and darwin only.
