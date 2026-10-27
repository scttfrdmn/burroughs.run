---
date: 2026-08-28
slug: the-suite-is-the-oracle
categories: [Building Burroughs]
wip: true
---

# 2. The suite is the oracle

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

## One rule, stated in five words

The project's engine law starts with this:

<!-- more -->

> **The suite is the oracle.** A feature exists when its spec tests pass. Claims of correctness cite counts (`N/N green`), not impressions.

The suite is the upstream WebAssembly spec test suite: `.wast` script files of modules and assertions about them, of which the core board reads 256. It is vendored into a gitignored path and pinned to a commit. The project never commits upstream test material. Burroughs runs the suite with its own pure-Go harness. ADR 0003 keeps non-Go tools such as `wabt` out of the conformance loop, so nothing outside the module decides whether the engine passed.

Running the harness prints the **board**, and the board is what everyone reads.

## Five ways to answer

Pass and fail turned out not to be enough. The board now has five columns, and each one exists because a two-column board was hiding something.

| Verdict | Meaning | Why it exists |
|---|---|---|
| `pass` | The engine gave the answer the suite expects | |
| `fail` | The engine gave a wrong answer | |
| `unsupported` | The harness has no case for this command at all | It measures what the harness *can ask*, not what the engine can do |
| `gated` | The vector needs a proposal whose gate is off | A feature behind a gate is neither passing nor failing, it is switched off |
| `unimplemented` | The harness can ask, but no engine component exists yet to answer | Added in ADR 0010 so that new questions don't inflate `fail` |

`unimplemented` came last. On 2026-08-01 the harness began admitting `(module quote …)` forms. These are modules given as text, so answering them needs a text-format parser. Admitting them raised the `unsupported` ceiling from **1345 to 26742**. ADR 0010 made two changes together *"because neither is honest alone"*: admit the form, and put its 1236 vectors under a new verdict rather than counting them as failures of a parser that didn't exist yet. The same day, ADR 0004 recorded 0010's guard 4, which made `unimplemented` a hard condition on releases: no minor version is cut while it is nonzero.

`gated` is the more important of the two. Every proposal beyond the core spec (SIMD, threads, GC, tail calls, exception handling, and so on) lives behind a **gate**, a feature flag in the decoder. Contract §9 G-1 says a gate may default on only when its own upstream suite is green. Behaviour 4 in the project brief tightens that: *"A flip is never in the mechanism's PR — it is its own stamp-tier event."* Code lands with the gate off. Turning it on is a separate decision with a pre-registered forecast of what the board will do, and it waits for a principal's approval.

## The reference interpreter as a second authority

The suite has a blind spot. Most of its vectors about opcodes are `assert_malformed`: they check that bad bytes are rejected. So a decoder that wrongly rejects a *valid* opcode is invisible on the board. Contract §9 G-3 calls this the accept-direction blind spot.

The repair was another authority. ADR 0007 made the upstream **reference interpreter** (written in OCaml) the source of truth for the opcode table. The table is *"machine-derived from, or machine-checked against, the reference interpreter. Hand-trusted is not on the menu."* No OCaml toolchain was assumed, so a pure-Go extractor reads the reference's `decode.ml` (542 arms, by the ADR's own correction of its first count of 504). The text format's keyword table was derived the same way from `lexer.mll` (ADR 0009).

The first extraction found defects that the board could never have shown. Grave #46 found four extractor bugs, including `i8x16.shuffle` reading one lane byte instead of sixteen. Grave #44 found **three controls that could not fail.** One was a check whose falsification test still passed with the fix removed, because a shorter pattern was masking the same text. That grave's lesson was *a control is not a control until it has been observed to fail on the defect it names.* A standing law in the same family, covered in the next post, came from Scott's ruling on PR #108: *a control isn't born until it has been watched die.*

## Board history

The early board grew fast. A few readings from commit messages and release notes:

| When | Reading | Source |
|---|---|---|
| 2026-07-30 | `binary.wast` alone: 49/127 pass, 78 fail | first harness commit |
| 2026-07-31 | 738 pass, 26 fail, 0 unsupported, 2 gated | PR #40's commit message |
| 2026-08-11, before the SIMD flip | 34308 pass / 118 fail / 27907 gated / 2689 unsupported | `v0.3.0` block |
| 2026-08-11, after the SIMD flip | 58429 / 279 / 3625 / 2689 | `v0.3.0` block |
| 2026-08-28, `v0.4.0` | **60957 pass, 0 fail, 0 unsupported, 4187 gated, 0 unimplemented** | `v0.4.0` block |
| 2026-10-26, `v0.11.0` | unchanged: 60957 / 0 / 0 / 4187 / 0 | `v0.11.0` block |

(The early rows count a smaller corpus. The board's corpus has been derived from capability rather than a hand-kept list since PR #54, so it grew as the harness learned to read more forms.)

``` { .text .fragment }
pass count, core board (not to scale between early points)

60957 |                                    ●━━━━━━━━━━━━━━━━━━━━━━━●  v0.4.0 … v0.11.0
58429 |                           ●  after SIMD flip
      |
34308 |                          ●  before SIMD flip
      |
  738 |  ●
   49 | ●
      +--+----+------------------+---------+----------------------+--
        07-30 07-31           08-11     08-28                  10-26
```

The flat line from August to October is not a stall. The `v0.6.0` notes say so: *"this release's work is in the component/async tier, which the core spec suite does not cover."* Later posts are about work the core board can't see.

## The first flip: SIMD

On 2026-08-11 `gate:simd` was turned on. This was *"the project's first default-behavior change."* Its own suite (59 `simd_*.wast` files) read **pass=25158 fail=161 gated=0**, so it was not green. Every one of the 161 failures was attributed, by the engine's own error taxonomy, to the validator (issue #9), which hadn't been built yet. Without it, any module reaching the interpreter unvalidated hit the same wall.

A literal reading of G-1, "green, full stop", couldn't be satisfied by any gate during v0. Scott stamped an amendment (ADR 0025, on #230) that carved out exactly that population and named its retirement condition. The carve-out ends *"when `ErrNotValidated` has no reachable call site in the engine — a state of the code, checkable by grep."* A later amendment (ADR 0043) noticed that an earlier version of the condition (*"retires itself when #9 lands"*) could be met by closing an issue without changing any code, and fixed it. That became a recurring kind of defect: a rule that a bookkeeping action could satisfy.

The flip's forecast was pre-registered and then reconciled: pass +24121, fail +161, gated −24282, unsupported +0, *"summing to zero against the fixed 65022 total — identical on both architectures."*

## Why there is no v0.1.0

ADR 0004's original version table tied minor versions to milestones: `v0.1.0` for the MVP core suite green, `v0.2.0` for the first gate flipped (GC), and one minor per gate after. SIMD flipped before GC, so `v0.3.0` shipped straight from `v0.0.1`. SemVer doesn't go backwards. When the MVP suite went green on 2026-08-28, the release was called `v0.4.0`, and its block opens:

> **This is the release ADR 0004's table numbers `v0.1.0`: the MVP core suite goes green.** The digit does not say so, and the meaning is written here rather than left to the number.

ADR 0004's 2026-08-28 amendment then separated the digit from the milestone mapping: *"the meaning now travels in the changelog entry."* Since then every release block says what it claims and also what it doesn't. That second part becomes important in post 9.

## v0 closes

`v0.4.0` recorded **all twelve of v0's closure conditions discharged** (#464, #499), the `v0 interpreter` milestone closed at **99 issues**, and two default-on gates (SIMD and relaxed SIMD). Behind the board numbers was a working interpreter: an execution loop, linear memory, the block family, `br_table`, calls and tail calls, globals, tables, reference types, exceptions, a WasmGC representation built on Go pointers the collector already traces (ADR 0020), and a text-format parser and encoder.

One gate was **declined for good**. `gate:endtable` would have resolved branch targets at decode time. It was measured at **+11.82%** on decode, against an execution saving of 3.4 to 4.8% on hot code, with break-even at roughly 126 dynamic block entries per instantiation. The flip was declined on those numbers. The release notes describe it as *"gated and permanently declined, which is a different state and is recorded as one."*

The 4187 gated vectors at `v0.4.0` belong to six core proposals that are still off by default: exception handling, memory64, GC, tail calls, multi-memory and extended constant expressions. A separate lane runs the board with every gate on. `TestAllGatesOnLeavesNothingGated` asserts that all 4187 are answered there. They are switched off, not unimplemented.

The project worried about what reaching 0 fail would do to it. One of its laws reads *"A zero-fail board is not a green light, it is a lost instrument."* It was minted, the law says, the usual way: *"4162/0 was reached and the next three PRs were instruments."* Failures grouped into buckets had been the work plan. Once nothing fails, the board stops saying what to do next, and the pull toward building more checks takes over. The next two posts are about that pull.
