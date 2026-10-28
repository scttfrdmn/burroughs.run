---
date: 2026-09-05
slug: the-process-that-built-itself
categories: [Building Burroughs]
written: 2026-10-26
---

# 4. The process that built itself, and the pruning

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

Post 3 described the governance machinery at its best. This one is about its tendency to grow. Every grave produced a law, every law could fail, and every way a law could fail looked like it needed a control. Left alone, a process built on "check the checker" will happily spend its time checking.

<!-- more -->

The project noticed. Several of its sharpest rulings are about stopping things.

## "A zero is a confession"

The first brake was built into the accounting. The brief's first behaviour says *the phase's product is the work; instruments are overhead on it.* v0's product was a running interpreter. A census, a board bound or a citation sweep is overhead that has to be *charged to* some piece of product work.

The rule was written down on 2026-08-02 (PR #97), four days in, and its commit message explains why: *"The board reached 4162 pass / 0 fail and the next three PRs were instruments."* That was the ratio from the last post going from 1:1.8 to 1:5.1. The commit adds a detail: *"internal/interp still holds 0 engine lines against 493 test lines."* At that point the interpreter package had tests and no interpreter.

For a while there was a hard stop: **two consecutive instrument-only PRs** halted work until a principal discharged it. That counter was retired on the review of PR #647 (2026-09-05), in Scott's words: *"the instrument/product counter is retired. Keep quoting the ratio; gate nothing on it."* The brief's explanation is short: the ratio prices the same drift on every PR without a threshold to argue over, and *"a figure nothing acts on cannot be gamed."*

## The brief that rationed itself

`CLAUDE.md` is the agent's brief, the file it reads at the start of every session. By mid-August it had become the laws themselves. On 2026-08-12 (#277) it was turned into an index and **46 laws were moved verbatim** into `docs/laws/`. Then, as you might predict, came a *size tripwire* on the index: a 38,400-byte budget (`claudeMDCeiling`), a per-entry byte ledger (`TestClaudeMDIndexLedger`), a golden file for the ledger, and a `make` target to regenerate it.

Five days later (2026-08-17, #377) Scott's four-workstream directive cut it back:

> CLAUDE.md inversion — cut it to a page of links, and retire the ceiling apparatus with it. [...] Once the file is a pointer page there's nothing to ration.

The retirement commit records each removed control and why, because even deletions follow a law here (*"a tripwire whose subject dissolves is re-pointed, never closed"*). Its note on the byte budget: *"a 38400-byte budget over a file that rations nothing measures a quantity nobody is trading."*

On the same day a second rule went. The brief had required every issue to have a milestone at filing time. Issue #324 counted **32 of 61 open issues** without one. Scott retired the rule rather than having the agent go milestone them all. The brief now says *"an issue attaches to one when it is scheduled, not when it is filed,"* because requiring it at filing *"prices filing at the cost of scheduling."*

## Decide and proceed

The #647 review made a second change, arguably bigger. Up to then the rule had been that anything Scott must decide is flagged and never decided on Scott's behalf. Combined with an agent inclined to see a decision everywhere, that sent a lot of questions upward. Scott narrowed it:

> Report decisions taken, not rulings requested. Escalate only for: contract (§) text, public API surface, reversing a stamped ADR.

Outside those three, the brief now says, *"a flagged question is **work not done**."* The three were chosen because each is something the agent cannot undo alone: normative text, something an embedder depends on, and another principal's stamp. The load-bearing section of every PR report became **Decisions taken**.

## The parked queue

The queue of decisions waiting for Scott (issues labelled `decision-needed:scott`) got the same treatment. On the #562 review Scott ordered it *parked*, meaning not reported at all, and gave a reason that's worth quoting for anyone who has written a status update:

> Until then, don't report the queue at all — "twelve waiting, none blocking" is itself the surfacing I asked you to stop, and half-parking pays both costs.

Parking needed an end point, and that's where things got tangled, in a way the project itself records with some humour. The end point was first "#10's §4 litmus battery". But the battery was itself parked past what spawn needed, so it couldn't serve as an end point. The lesson the brief minted: **"a parked subject cannot be another subject's end point."** The end point moved to "spawn runs". Spawn ran, and Scott moved it again, to "v1's threads are complete": join, #586's resolution, and the battery.

Then on the #670 review Scott took the `gate:threads` flip off the schedule:

> those six rows have exactly one consumer, and it's a condition I invented … The condition stands, but the flip isn't scheduled — which means the battery isn't either.

That made the parking's end point unreachable again, the same lesson *"landing for the **third** time."* The brief declines to repair it on its own initiative, because re-pinning the end point is Scott's move.

The brief is candid about its own drift. One paragraph begins: *"v1's artifacts have started, and this paragraph said they had not for as long as that was false."* The project calls this pattern the **foreclosing-words shape**: a sentence written before a change, left standing after it, telling the next reader the tree is in a state it isn't in.

## Processes that outlive their purpose

Some pruning was about actual processes. One session launched several CI watchers, relaunched them after the harness killed one, and at one point had two watching the same run. Scott's rule on the #658 review: **no process starts without naming what ends it.** The repair (ADR 0072, `scripts/detach.sh`) records the pid and process group *before* starting, bounds the run by a wall-clock timeout and the session's liveness, and writes a terminal line on every exit path.

Even the repair had a bug in its own class. Killing the process group also killed the wrapper that records status, so a naive loop polled for the whole deadline over a child that no longer existed: *"the leak class reproduced inside its own repair."* Later, on 2026-10-25 (#1105), it turned out the script used `BASHPID`, which doesn't exist in macOS's stock bash 3.2. On that shell the wrapper died before running anything.

The brief's waiting rule has its own history: *"`sleep` is never how you wait for a signal that exists."* CI is watched from the run for the pushed SHA, because `gh pr checks --watch` *"races the run's creation and reports the previous commit's green."*

## `make check` is not `make ci`

On the #829 review, `make strict` turned CI red on a tree that `make check` had just passed. The response (#832, 2026-09-30) split the gates. `check` is the fast inner loop and **`ci` is what must be green before any report.** `check` now prints, on success, which gates it did *not* run, so its green can't be read as a green tree. A control (`TestCIGatesCoverWhatCIInvokes`) derives CI's set of targets from the workflow file. It was added because the first hand-written list *"named five targets CI does not invoke and omitted three it does."*

## Was it too much?

Partly, and the project says so in its own records. The brief carries a whole section explaining why a lab-hardware table copied from another page is exempt from the brief's own rule against measured figures. The rule that agent worktrees live outside the repository stays a line of prose rather than a control, because a control *"would have to run inside the tree it is checking for copies of itself."*

The pruning shows the system working, though. Each cut was made by a principal, recorded with its reason, and kept the useful part: the ratio outlived its counter, the laws outlived their index, and the escalation rule outlived its open-ended form. The process could not stop itself from growing, but it kept a record good enough for Scott to see where to cut.

The next posts go back to the engine, where the work got hard.
