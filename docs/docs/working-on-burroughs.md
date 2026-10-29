---
wip: true
---

# Working on Burroughs

*For contributors: where the project's records are, how conformance is asked, and how the gates run. This page is
drawn from the repository as of `v0.12.0` and its main branch on 2026-10-28. The repository's own files win where
they disagree, and the [contract](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md)
is normative.*

## Reading the repository

- **[`docs/burroughs-contract-v0.1.md`](https://github.com/scttfrdmn/burroughs/blob/main/docs/burroughs-contract-v0.1.md)**
  is the normative host contract. Start there.
- **[`docs/decisions/`](https://github.com/scttfrdmn/burroughs/tree/main/docs/decisions)** holds the accepted
  decision records (ADRs). A design choice gets one before it gets code, and each record's `Status:` line cites the
  approval it rests on.
- **[`docs/laws/`](https://github.com/scttfrdmn/burroughs/tree/main/docs/laws)** holds the project's disciplines,
  grouped by family. Each law carries its specimen and the finding that minted it.
- **[`CLAUDE.md`](https://github.com/scttfrdmn/burroughs/blob/main/CLAUDE.md)** is the implementation agent's brief:
  the phase ladder, where work is tracked, and the behaviours that change what a pull request does. It is a pointer
  page, not the corpus.
- **[`CHANGELOG.md`](https://github.com/scttfrdmn/burroughs/blob/main/CHANGELOG.md)** says what landed, newest first.
  Gate flips are **Added** and graves are **Fixed**, each beside its issue.

**The README's examples are executed.** `TestREADMETranscriptIsExecutable` runs every `burroughs` line in the README's
transcript through the CLI's own dispatcher and compares the output and the exit code. `TestREADMEGoBlocksAreRealCode`
requires every Go block there to appear in compiled code. `examples/add/add.wasm` is committed so a fresh clone needs
one command, and it is derived: `go run ./examples/add/gen.go` re-assembles it from `add.wat` with the engine's own
assembler, and `TestExampleWasmIsDerivedFromItsWat` holds the two equal.

**Project state lives in GitHub, not in files.** Milestones are the phase ladder, issues are the units of work, and
issues labelled `type:grave` are the graveyard: bugs found, each with the lesson in its closing comment.

## The gates

- **`make ci`** is the gate a change must pass before it is reported. It runs `make check`'s gates plus the ones
  `check` does not: the race detector, the no-skip run (`strict`), the opcode and keyword drift checks, the
  witnesses, the citation, closing-keyword and spacing sweeps, and the vulnerability scan.
- **`make check`** is the fast inner loop: format check, build, vet, lint, the tests (default and end-table builds)
  and dead-code. On success it prints the gates it did **not** run, so its green cannot be mistaken for `ci`'s.
- **The toolchain is pinned exactly**, by `go.mod`'s `toolchain` line. The Makefile exports `GOTOOLCHAIN` from it,
  and `make toolchain-check` fails if the version that actually ran is a different one, so a green gate means green
  on the pinned toolchain.
- A test named on the Windows lane's [triage map](https://github.com/scttfrdmn/burroughs/blob/main/.github/windows-triage.tsv)
  is a known portability gap with an open issue. Windows is measured and not supported, and a failure the map does
  not name is new, so it is classified and filed before the next merge.

## Conformance

The upstream WebAssembly spec test suite is the oracle. The project quotes no figure from it in prose, because a
count typed into a document rots. Ask the instrument instead:

``` { .console .fragment }
$ make spec-tests                                    # vendor the suite (gitignored)
$ go test ./internal/spec/ -run TestPhase1Files -v   # print the board
$ make conformance                                   # assert it, both lanes
```

The board prints per-file counts, a total, and the failure buckets in size order. There are five verdicts, and the
distinctions between them are the point (decision 0010):

- **pass:** the engine answered, and the answer was right.
- **fail:** the engine answered, and the answer was wrong. This column is the work plan.
- **gated:** the engine could answer, but a proposal's gate is off in this build. This is absence by configuration.
- **unimplemented:** the harness asked, and the engine has no component to answer with. This is absence by
  construction. The board's bound on it is zero.
- **unsupported:** the harness cannot ask, because no command kind recognizes the form. This bound is zero too. A
  zero here is a fact about the harness's vocabulary, not an engine milestone.

`make conformance` runs two lanes: the default gate set, and an all-gates-on lane where the `gated` count must be
zero. It refuses to run without the vendored suite, because a suite that skips passes by asking nothing.

## Shared lab hardware

Everything above runs locally with nothing but a Go toolchain. Cross-architecture checks and recorded benchmarks run
on shared lab machines, which carry several projects at once, so **every run on them goes through that host's
[pueue](https://github.com/Nukesor/pueue) queue.** Contention then waits, instead of moving somebody's numbers.

- **`measured`** has one slot, for anything whose timing is recorded, compared or committed.
- **`build`** is wide, for compiles, unit tests and lint. Its results are not measurements.

A group excludes only its own tasks. A long `build` job therefore shares the machine with a `measured` one, and a
recorded figure must also check the host's other groups. A two-arm benchmark is submitted as one task, so another
project's job cannot land between its rounds. The host, group, task id and how busy the box was are recorded beside
the numbers.
