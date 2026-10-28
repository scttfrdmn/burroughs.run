---
wip: true
---

# Tutorial 3: load a component and call typed exports

*The program and its output are this site's `examples/component/` (`main.go`, `output.txt`). It uses fixtures
committed in the Burroughs repository under `internal/component/testdata/`; CI fetches them at the latest release
tag into `examples/component/testdata/` (the list is `examples/fixtures.txt`), runs the program with `go run .`, and
fails if the output differs from what is shown here. To run it yourself, copy those files from the engine repository
at the tag into `testdata/` beside `main.go`.*

A **component** is a WebAssembly artifact whose exports have WIT types (`u32`, `string`, `list<T>`, records, and so
on) rather than core `i32`s. Burroughs runs components by default (`gate:components`, since 2026-09-11) and gives you
two ways in:

- `LoadComponent` returns a `*Component` whose exports you call with `ComponentValue`s.
- `ComponentConfig.Run` runs a component that exports `wasi:cli/run`, the component-world equivalent of `_start`.

## The fixtures

| Fixture | Exports used | Signature (from its `.wat`) |
| --- | --- | --- |
| `record-synth.wasm` | `addrec`, `makerec` | `addrec: func(p: record { a: u32, b: u32 }) -> u32`, `makerec: func() -> record { name: string, n: u32 }` |
| `string-echo.wasm` | `run-sync` | `func(s: string) -> u32`, the length of the string after a round trip through a child component |
| `string-arg-realloc-synth.wasm` | `sum-list` | `func async (xs: list<u32>) -> u32`, an async-lifted export (`gate:async`, on by default) |
| `trap-then-call-plain.wasm` | `run`, `ok` | `run` traps (`unreachable`). `ok` returns 7. |
| `p3hello.wasm` | `wasi:cli/run` | prints `Hello, world!` (a `wasi:cli@0.2.3` world) |

Each fixture's `.wat` source sits beside it (except `p3hello`, whose `.wit` is there instead), if you want to read
what the guest does.

## The program

```go title="examples/component/main.go"
--8<-- "examples/component/main.go"
```

## Output

Run with `go run .` in `examples/component/`:

```text title="examples/component/output.txt"
--8<-- "examples/component/output.txt"
```

!!! note "Fixed in `v0.12.0` ([#1129](https://github.com/scttfrdmn/burroughs/issues/1129))"
    The line `wrong argument type at the call` shows a mistyped `Component.Call` argument refused as **the caller's
    error**: the message names the parameter and both types, and it matches no sentinel, as a mistyped
    `Instance.Call` argument does. In `v0.11.0` it was classed `ErrUnsupported`, wrongly blaming the engine. This
    note said the line would change when the fix shipped and that this site's CI would flag it, and both happened.

## What to take from it

- **Types are part of the value.** Compound values need a `ComponentType`, because an empty `list<u32>` and an
  empty `list<string>` are different values. Most mistakes are refused when you build the value, before any call.
- **A record is built from a map, and its field order comes from the type.** Declare the fields in the order the
  component's WIT does.
- **A type mismatch at the call is your error, not the engine's.** It matches no sentinel, so `errors.Is` against
  `ErrUnsupported` is false, and the message says which parameter and which types.
- **Read string contents with `Str()`.** `String()` is a debug rendering: `string("burroughs")`.
- **A trap locks the whole component instance.** The next call returns `ErrInstanceLocked`, which wraps the same
  `*Trap`. Core modules do not do this. To recover, load a new `Component`.
- **Close what you load.** `Close` is idempotent. A call after it returns `ErrComponentClosed`.
- **`ComponentConfig`'s nil streams discard output.** The second `p3hello` run printed nothing anywhere.

## Limits you will hit

- **Only `u32`, `string`, `list` and `record` cross.** A function taking or returning `bool`, `u64`, `s32`, `f64`,
  `char`, `variant`, `enum`, `option`, `result`, `flags` or a resource handle is refused **by name** at the call,
  as `ErrUnsupported`.
- **The WASI 0.2 host is partial.** The component host provides stdout and stderr streams, stdin reads, arguments,
  `exit`, and error debug strings. `wasi:cli/environment.get-environment` and
  `wasi:filesystem/preopens.get-directories` both return **empty lists**. Clocks, random, sockets and http are not
  provided. `ComponentConfig` has no `Env` or `Preopens`.
- **Calls into one `Component` run one at a time**, interleaving only where a call suspends. Load several
  components for parallelism.
- **Component-model threads** (`gate:cm-threads`, phase 1 of the shared-everything-threads proposal) start
  through these paths, cooperatively, one at a time per instance. The proposal's `thread.spawn-ref` is refused by
  name, and `thread.spawn-indirect` is refused at decode.
- **From the CLI, a component takes its arguments after `--`**, and refuses `--dir`, `--scratch`, `--env` and a
  function name by name (since `v0.12.0`, [#1128](https://github.com/scttfrdmn/burroughs/issues/1128)): see the
  [CLI reference](../cli.md#components-through-the-cli).
