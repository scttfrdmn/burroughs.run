---
wip: true
---

# Tutorial 4: traps, refusals, gates and the environment switches

*The program is this site's `examples/refusals/` (`main.go`, the `.wat` sources and their assembled `.wasm`,
`output.txt`, `output-threads0.txt`), and the CLI transcripts are `examples/cli/`. CI runs all of them against the
latest Burroughs release, with the CLI built at that tag, and fails if any output differs from what is shown here.
`[exit N]` in a transcript is the exit status when it is not 0.*

Burroughs avoids returning one undifferentiated error. "Fix your module", "rebuild with that gate on" and "this
engine is incomplete" ask different people to do different things, so each has its own classification. This tutorial
builds one module per outcome and shows how each one is reported. [Concepts](../concepts.md#the-vocabulary-trap-refused-declined-unsupported-gated)
covers the vocabulary itself.

## The modules

| File | Contents | Expected outcome |
| --- | --- | --- |
| `malformed.wasm` | the bytes `not a wasm module` | malformed |
| `invalid.wat` | `(module (func (export "f") (result i32) i64.const 42))` | invalid: type mismatch |
| `memory64.wat` | `(module (memory i64 1))` | gated: memory64 is off |
| `tailcall.wat` | a function using `return_call` | gated: tail calls are off |
| `imports.wat` | imports `env.log` | unlinkable: the core path supplies no imports |
| `shared.wat` | a shared memory, `i32.atomic.rmw.add`, `memory.atomic.wait32` | runs: the threads grammar is on |

`shared.wat`:

```wat title="examples/refusals/shared.wat"
--8<-- "examples/refusals/shared.wat"
```

Assemble the `.wat` files with `wasm-tools parse X.wat -o X.wasm` (the site commits the assembled `.wasm` beside
each source).

## The program

```go title="examples/refusals/main.go"
--8<-- "examples/refusals/main.go"
```

## Output, default build

Run with `go run .` in `examples/refusals/`:

```text title="examples/refusals/output.txt"
--8<-- "examples/refusals/output.txt"
```

Notes on each line:

- **Malformed and invalid** carry the spec's own text, so you can match on `type mismatch` without Burroughs
  restating the spec's rules.
- **Gated** names the proposal. The module is well-formed, and a build with that gate on would run it. **The public
  API cannot turn memory64, tail calls, GC, exception handling, multi-memory or extended-const on.** These six are
  off by default and no `Feature` reaches them, so through the public API this result is final in `v0.11.0`.
- **Unlinkable**: the core `Instantiate` path has no way to supply imports, so the module is refused when it loads,
  before anything can trap on a missing function. If your module needs host functions, the only supported host
  imports are WASI preview 1 (`WASIP1Config.Run`) and the WASI 0.2 subset for components.
- **`Strict: true` changed nothing here.** None of these modules declines. No module in the default feature set is
  currently known to produce a decline. `Strict` only matters once the validator meets a construct it has no rule
  for yet.
- **Shared memory runs**, because `gate:threads` is on by default. There is only ever one agent,
  though: nothing public can start a second thread, so `memory.atomic.wait32` can only return 2 ("timed-out"), or 1
  ("not-equal") if the value already differs.

## Traps

A trap is not a refusal: the module was accepted and ran, and the program then went wrong. It comes back from
`Call` (or `Run`) as a `*burroughs.Trap`, whose `Reason` is the spec's text:

```console title="examples/cli/transcripts/t4-trap.txt"
--8<-- "examples/cli/transcripts/t4-trap.txt"
```

A core instance is still callable after a trap. A component instance is **locked** by one: every later
`Component.Call` returns `ErrInstanceLocked` wrapping the trap ([tutorial 3](03-components.md)).

## The same modules through the CLI

```console title="examples/cli/transcripts/t4-cli.txt"
--8<-- "examples/cli/transcripts/t4-cli.txt"
```

## The environment switches

Each default-on gate has a rollback: set its variable to `0`. Any other value, or leaving it unset, keeps the
default. The variables are read when they are used, so they work for the CLI and for your own program alike.

| Variable | Gate | Default since |
| --- | --- | --- |
| `BURROUGHS_THREADS` | `gate:threads`: shared memories, the 0xFE atomics, wait/notify | v0.10.0 (ADR 0106) |
| `BURROUGHS_COMPONENTS` | `gate:components`: the component model | 2026-09-11 |
| `BURROUGHS_ASYNC` | `gate:async`: the async component ABI | 2026-09-18 |
| `BURROUGHS_CM_THREADS` | `gate:cm-threads`: component-model threads, phase 1 | 2026-10-17 (v0.8.0) |

SIMD and relaxed SIMD are also on by default, but have **no** environment rollback.

### `BURROUGHS_THREADS=0`

The same program with the gate rolled back (`BURROUGHS_THREADS=0 go run .`) prints the same lines, except for the
last module:

```text title="examples/refusals/output-threads0.txt"
--8<-- "examples/refusals/output-threads0.txt"
```

With the gate rolled back, a `wasip1` guest that needs atomics can still ask for them with `FeatureThreads` (CLI:
`--features=threads`). `sharedcmd.wasm` (source in `examples/refusals/sharedcmd.wat`) is a tiny command with a shared memory that calls `proc_exit`:

```console title="examples/cli/transcripts/t4-threads.txt"
--8<-- "examples/cli/transcripts/t4-threads.txt"
```

(`sharedcmd.wasm` exits with the value the atomic add returns, the old contents of address 0, which is 0.)

`Config`, the core-module path, has no `Features`. With `BURROUGHS_THREADS=0`, a core module that uses threads
cannot be admitted through `Instantiate` at all. The last command above shows the CLI's core path refusing it even
with `--features=threads` (exit 6), because `--features` has no effect on that path
([#1128](https://github.com/scttfrdmn/burroughs/issues/1128), on the `v0.12.0` milestone, expects such a flag to be
refused by name instead).

### `BURROUGHS_COMPONENTS=0` and `BURROUGHS_ASYNC=0`

```console title="examples/cli/transcripts/t4-gates.txt"
--8<-- "examples/cli/transcripts/t4-gates.txt"
```

Both exit 6, and `errors.Is(err, burroughs.ErrGated)` holds for them in Go. (`p3hello.wasm` and
`p3async-hello.wasm` are in the engine repository's `internal/component/testdata/`, fetched at the tag. Both print
`Hello, world!` with the gates at their defaults: the last command above, and `p3hello` in the
[CLI reference](../cli.md#transcripts).)

## An unknown feature name

```console title="examples/cli/transcripts/t4-features.txt"
--8<-- "examples/cli/transcripts/t4-features.txt"
```

That is `ErrUnsupported`, exit 5. An unrecognized name is refused rather than ignored.

## A branching template

`describe` is this site's `examples/describe/`, which CI runs; the function as it appears there:

```go title="examples/describe/main.go"
--8<-- "examples/describe/main.go:describe"
```

Its `main` runs `describe` over one error of each kind it can produce without fixtures:

```text title="examples/describe/output.txt"
--8<-- "examples/describe/output.txt"
```

**Order matters in one place: test `ErrInstanceLocked` before the `*Trap` case.** A locked-instance error wraps the
trap that caused the lock, so it also satisfies `errors.As(err, &trap)`. If the trap case came first, the locked case
could never be reached.
