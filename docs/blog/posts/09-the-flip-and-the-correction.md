---
date: 2026-10-26
slug: the-flip-and-the-correction
categories: [Building Burroughs]
written: 2026-10-26
---

# 9. The flip and the correction

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

## A condition, quoted

For most of the threads phase, `gate:threads` sat behind a condition Scott had set, which the brief quoted word for word: the default **flips when #586 is resolved and the §4 battery covers what the default would then promise.** The brief adds that *"is a condition, not a deferral — the difference being that a reader can check whether it holds."*

<!-- more -->

By 2026-10-25 both halves held. #586 (a thread left on an abandoned memory image) had been resolved by ADR 0073 since September. The battery had no registered hole after #1100 (post 6). The chair ruled to *"prepare the `gate:threads` flip, decided unless Scott objects, after G1 is merged and green."*

## G1: proving a flip changes nothing it shouldn't

The forecast named its first gap, G1. If threads were on by default, would any *non-threads* module behave differently? A board total that stays the same isn't enough to show that, because a gain in one place can hide a loss in another. So G1 (`TestThreadsGateLeavesTheCoreBoardUnchanged`, #1102) compares **all 256 core files with the gate off and on, command by command**, and requires them to read identically. Only the threads proposal's own files may differ, and they must.

## The flip

ADR 0106 (2026-10-25) chose between three options: leave the default off; flip with no rollback but a revert, as SIMD had; or **flip with an environment-variable rollback**. It chose the third, because unlike SIMD, threads have runtime behaviour an operator might need to switch off in the field. The whole change is one line:

``` { .go .fragment }
func DefaultFeatures() Features {
	return Features{SIMD: true, RelaxedSIMD: true, Threads: os.Getenv(threadsGateEnv) != "0"}
}
```

It was built and held as PR #1103. Scott's answer came the same day, relayed by the chair and recorded on #10: *"I do not object."* It covered the flip, the claim text, and the `v0.10.0` release.

Two details of the ADR are worth noting:

- **The off state stays witnessed.** The tests that pinned the old default now run under `BURROUGHS_THREADS=0`, so the rollback is shown to restore exactly the refusal it promises. Each has an on-state arm beside it. In the ADR's words: *"A gated rule's off state is its own claim."*
- **A stock Go guest never needed this.** `TestTheGoGuestNeedsNoThreadsFeature` asserts that an ordinary `wasip1` Go program decodes with threads off (post 7). Guests built by the Phase 4 fork are the ones that need the gate.

`v0.10.0` also added a third gating CI leg, `macos-15` on arm64. It was added on the chair's second ruling of 2026-10-25, and the release note gives the reason: *"so the claim's platform line matches the CI that gates merges."* Before that, darwin evidence came from the development host's own runs, *"which gate nothing."*

## The claim

The `v0.10.0` threads claim, as amended by the chair and not objected to by Scott:

> **On linux and darwin**, Burroughs runs WebAssembly threads on OS threads in pure Go: shared memories, the threads proposal's atomics, and `memory.atomic.wait`/`notify` backed by a futex-style wake. Blocking host calls block only their own thread. The boundary memory model of contract §4 holds at every host↔guest crossing the engine has.

It was followed by the careful evidence breakdown from post 6: 17 of 19 cases, which cases were judged by `-race` and which by scheduling, *"No case is judged by hardware weak-memory values."*

## The next day: what the claim didn't mean

On 2026-10-26, under the chair's fourth ruling of the day, a narrower question got its own test: **can anything a user can reach actually start a second thread over a shared memory?** PR #1121 checked every public entry point:

| Public path | The guest's way to spawn | Result |
|---|---|---|
| `Instantiate` | the fork's `"burroughs" "spawn"` import | link refused, naming the import |
| `WASIP1Config.Run` | the same, in a command | link refused, naming the import |
| `burroughs run` | the same, as a module and as a command | non-zero exit, naming the import |
| `LoadComponent`, `ComponentConfig.Run` | `thread.spawn-ref` | refused by name |
| same | `thread.spawn-indirect` | refused at decode, as "malformed" (itself a misclassification, #1120) |

The public API also has no spawn, join, stop or host-function registration. So:

> a public module's shared memory has exactly one agent, and a `memory.atomic.wait` on it can only time out.

The engine runs threads, and the litmus battery measures that by driving the engine's internal API. `gate:threads` admits the threads *grammar* everywhere. But between the two there was no door. The PR's figure says it in numbers: *"public entry points with a test answering whether a guest can start a second shared-memory thread: 0 of 5 before, 5 of 5 after. The answer on all five is no."* Its ratio line reads `engine 0 / instrument 185 = 1:inf`, which is the honest result for a PR that only measures.

The PR also recorded what *does* work, so the correction couldn't be read as "no threads of any kind": component-model Threads (`gate:cm-threads`) do start through the public component path. They run cooperatively, one at a time per instance. They are not agents over a shared memory.

## Correcting a release

The project's records discipline made the correction straightforward:

- **The `v0.10.0` tag and its changelog block are unchanged.** A release note records what was claimed at the time.
- **`v0.11.0`'s block leads with the correction.** After its default-change line (*"What changes by default: nothing."*), its first paragraph opens *"The claim, corrected first."* The correction is also posted to #10 and to the `v0.10.0` release issue (#1108), following the law that *"a correction travels attached to the quote."*
- **The fix is not in the correction.** Exposing a threads API is public-API surface, one of the three things the agent may not decide alone. So it went to Scott as a question. Scott's answer (relayed in the chair's addendum of 2026-10-26) became `v0.12.0`'s goal:

> Guest started seems to be the intent of the project - Go->Wasm->Burroughs runs a Go program in all its glory on an engine that can.

Host-started threads are not being built.

## v0.11.0, in full

`v0.11.0` (2026-10-26) is a modest release with an immodest amount of honesty in it:

- **"What changes by default: nothing."**
- The correction, first.
- **The battery at 19 of 19**: SP-5 and H-4 via the component-model `parked-runner` fixture, 40,000 `-race` iterations with no failure (post 6).
- **The oracle pinned by checksum** (post 7).
- **Fixes:** #1106's lost wakeup, #896's bound set from measurement, grave #1118's witness (post 8), and `detach.sh` running under macOS's bash 3.2.
- **Six known limits**, each with an issue, starting with *"No public path starts a second thread over a shared memory (#1122)."*

## After v0.11.0: one lane per gate

One more instrument landed on main after the tag (#1127, in `[Unreleased]` at the time of writing). The all-gates-on lane says the six default-off core gates *together* leave nothing failing. A flip forecast needs to know what turning on *one* gate does. `TestEachDefaultOffGateOnItsOwnLane` finds every default-off gate by reflection, so a new gate gets a lane without editing the test, and runs the board with that gate alone. It was read on darwin/arm64 and janus x86-64, *"identical in every counter,"* and the forecasts were posted for Scott's stamp. No gate flips there. Each flip is still a separate decision.
