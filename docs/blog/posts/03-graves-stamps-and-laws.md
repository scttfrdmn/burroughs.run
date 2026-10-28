---
date: 2026-08-28
slug: graves-stamps-and-laws
categories: [Building Burroughs]
written: 2026-10-26
---

# 3. Graves, stamps and laws

*Drafted 2026-10-26 from the repository's records. Corrections welcome.*

Most of Burroughs' record is not code. On 2026-10-26 the tree held 106 decision records, eleven "law" families of about 5,600 lines, and a changelog of about 19,500 lines. The tracker held 168 issues labelled `type:grave`, plus 4 PRs carrying the same label. This post is about why an engine project produced that much text. Some of it is very good practice. Some of it, as the next post shows, got out of hand.

<!-- more -->

## The setup: an agent that writes the reports

The arrangement matters. The implementation agent (CC) writes nearly all of the code, tests, ADR drafts and reports. Scott owns the project and makes the decisions, and chat-Claude reviews the architecture. The two principals see the work mostly through what the agent writes about it.

That makes honesty in reporting the central engineering problem, not a side issue. If the agent decides what counts as done, chooses the measurement, and writes the report, every weakness in its self-assessment goes straight into the project's record. Much of the machinery below is an answer to the question: *how do you trust a report written by the actor being judged?*

The answer was written down early (Scott's compression, PR #113):

> **The actor never chooses the instrument that judges the actor.**

## Decision before code

Design choices get an ADR (`docs/decisions/NNNN-*.md`) with context, options, the choice and its consequences, and they get it *before* implementation. Discussion happens in an issue labelled `type:decision`. The ADR is described as *"the tombstone, an accepted record only."*

ADR titles here are long sentences, because each title carries the decision and its reason. A few examples:

- *0019 — runtime type tests are events; subsumption is #9's, whenever #9 exists*
- *0059 — the safepoint poll is guarded at the pc assignment, because a back-edge is a runtime comparison and straight-line code pays nothing*
- *0102 — the pinned reference defines the semantics; a toolchain's stricter refusal is followed unless Burroughs implements and witnesses what the pin admits*

The strictest rule concerns the `Status:` line:

> A status field is a citation to an approval … *an ADR marked accepted on a stamp nobody gave is a fabricated citation about the project's own governance* — worse than a wrong option, since a wrong option is arguable and a forged provenance is not. (Ruling: Scott, PR #142.)

So an ADR stays `proposed` until a stamp exists to point at, and the record keeps both the stamp and how long the ADR waited for it. This is enforced in practice. In `v0.11.0`, ADR 0084 was corrected to `accepted` because it had read `proposed` *"for seven weeks after"* Scott's stamp. ADRs 0102 and 0076 were left `proposed` because *"no stamp for either is on record."*

There's also a cap on paperwork: **one ADR earns one implementation.** An ADR whose code hasn't started is a reason to write code, not another ADR.

## Graves

A **grave** is a closed issue labelled `type:grave`, with the lesson in its closing comment and a comment at the fix site citing the issue number. The point is that the lesson outlives the bug. Lessons are indexed by *shape*, not by file, so a defect that feels familiar has probably been paid for already, and a new grave triggers a sweep backwards through its family.

The first ones were in the genesis commit (post 1). By late October 168 issues carried the label, plus 4 PRs. Some titles:

- #18: *a lexer's delimiter set can be right for the grammar and wrong for the corpus*
- #32: *a test asserted a property of code that doesn't run yet*
- #44: *a control that cannot fail — three instances in one session*
- #101: *Instr.Op was a byte, truncating the 0xfd region's high opcodes into other instructions*
- #1118: *the one-load witnesses assumed A was still filling when B grew, and failed main against a correct engine* (post 8)

Graves also caught failures in the process. Grave #314 found that GitHub's `Closes #N` keyword closes an issue on merge before anyone writes the closing comment with the lesson. That leaves *"a tombstone with no inscription, which is worse than an open issue."* The law that followed explains why the keyword is now banned outright: GitHub's parser *"reads tokens, not negations, so no sentence around the keyword can disarm it: 'this does not close #N' closes #N."*

## Controls, and the controls on controls

A **control** here is a test whose job is to keep another claim true: a census of gates, a floor under the pass count, a check that every link in the docs resolves. The project's key insight about controls is grave #44's lesson, in the form of the law that states it:

> **A control isn't born until it has been watched die.** … A no-op guard and a working guard produce identical output on every input that doesn't trip them, which is all of them. (Ruling: Scott, PR #108.)

So every control ships with a **neuter**: deliberately break the thing it guards, run it, and watch it fail. CHANGELOG entries say this consistently, in the form *"Neuter, run: with the comparison removed, the substituted archive installs."* A control whose neuter passes is classified as *"stillborn,"* and the law calls that *"the most valuable outcome the exercise has."*

The law then grew a three-way classification for when the falsification passes: a mutation that changed nothing (one "swap" of two calls had only renamed a variable), a blind control, or a wrong prediction by the author. The third is described as *"the battery's highest function, because it falsifies the author's own model of the control."*

The laws themselves are under control too. Three tests in `internal/testenv` keep the corpus from rotting: every markdown link in the tree resolves to an existing file and heading, no link is broken across a line, and every law family is reachable from the project brief. They are three separate tests because, per grave #34, they fail for unrelated reasons.

## The PR is the report

Every piece of work happens in a PR, including self-merged ones, and the PR description is the session report. It has fixed sections: **Board** · **Landed** · **Decisions taken** · **Decisions needed from Scott** · **Graves** · **Done since last review** · **Next**.

The last-but-one section has a story. Scott added "Done since last review" on the #387 review after asking three times about work that was already on main: *"I asked three times for something that was already on main, which is my time and yours spent on a lookup you'd already done."* A report that only looks forward makes the reader do the lookup.

Each PR also states two figures:

1. **The `unsupported` delta, where "a zero is a confession."** If a PR didn't change what the harness can ask, it says so in that word and names the product work it was overhead *for*.
2. **The instrument-to-engine ratio**, produced by `make ratio` and pasted from the tool, never typed by hand. "Instrument" is tests, generators and harness code. "Engine" is code in the module path. The figure is quoted and never compared to a threshold.

The ratio rule came from an embarrassing measurement. Over six merges the ratio went from **1:1.8 to 1:5.1** while the `unsupported` column didn't move at all. The law calls that *"the definition of spinning, and it was invisible because every individual PR was defensible."*

Each commit is meant to carry a provenance trailer: `Ratio-Class: ordered <citation>` if a principal ordered the work with a citable artifact, `Ratio-Class: carried` otherwise. A commit without one is counted `unattributed`. Many ADRs carry a line like *"Recorded by the actor the ruling was given to, which is not independent provenance — durability is not independence."* The agent writing a ruling into the repository doesn't make the ruling independently verified, and the record says so each time.

## Testimony

A whole law family (`errors-and-testimony.md`) treats error messages, comments and ADRs as **testimony**, meaning things that can lie. The decoder's error strings copy the spec suite's verbatim. A comment that describes code wrongly counts as a defect even if the code is right. Where prose and an executable reference disagree, *"the executable outranks."*

ADR 0002's amendment (post 1) shows the attitude: the falsified sentence stays in place, marked and explained, because removing it removes the evidence. The `v0.4.0` release note's *"it is not v1"* was later overtaken by events and got the same treatment. It was kept as written, with a superseded box pointing to ADR 0099.

## Does it work?

By the evidence in the tree, mostly yes, at a cost.

- The board went from 49/127 on one file to 60957/0 across the core suite in four weeks. Defects stayed visible because a fail could not be quietly relabelled.
- Release notes list their own known limits and corrections first. Post 9 describes a release whose top line is a correction of the previous day's claim.
- Every concrete figure in these posts could be traced to a specific commit, issue or release block. That is unusual.

The cost is volume. A 19,500-line changelog for a three-month-old project isn't something people read; it's something you search. The process also kept growing new layers to watch the old ones. The next post is about the times Scott looked at what had grown and cut it back.
