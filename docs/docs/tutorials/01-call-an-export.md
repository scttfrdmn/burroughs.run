---
wip: true
---

# Tutorial 1: call an exported function from WAT

*The program and its output are this site's `examples/embed-core/` (`main.go`, `calc.wat`, `output.txt`). CI runs
it with `go run .` against the latest Burroughs release and fails if the output differs from what is shown here.*

You will write a small module in the WebAssembly text format, assemble it, and drive it from Go: list its exports,
call functions with typed arguments, read multiple results, catch traps, and read a global.

## 1. The module

`calc.wat`:

```wat title="examples/embed-core/calc.wat"
--8<-- "examples/embed-core/calc.wat"
```

The module has **no imports**. That matters: the core `Instantiate` path supplies none, so a module that imports
anything is refused with `ErrUnlinkable` (see [tutorial 4](04-traps-and-refusals.md)).

Assemble it. Burroughs reads binaries only:

``` { .console .fragment }
$ wasm-tools parse calc.wat -o calc.wasm
```

## 2. The module file

```text title="examples/embed-core/go.mod"
--8<-- "examples/embed-core/go.mod"
```

It requires the published `v0.11.0` from the Go module proxy; there is no local `replace`. (CI points the `require`
at the latest release tag before it runs the example.)

## 3. The host

`main.go`:

```go title="examples/embed-core/main.go"
--8<-- "examples/embed-core/main.go"
```

## 4. Run it

Run it with `go run .` in `examples/embed-core/`. It prints:

```text title="examples/embed-core/output.txt"
--8<-- "examples/embed-core/output.txt"
```

## What to take from it

- **`Exports()` is the set of names `Call` accepts**, in declaration order. Only functions are listed.
- **Types are checked at the boundary.** An `I64` where the signature says `i32` is refused before any frame is
  built. These caller errors are plain errors and match no `burroughs.Err*` sentinel.
- **A trap is a value you can match.** `trap.Reason` is the spec's own text. The core instance stays usable
  afterwards: `add` still ran.
- **`calls` is 3**: the two `add` calls and the one `mul64` call that ran. The calls refused at the boundary never
  executed, so they did not count.
- **A result's `String()` is a valid CLI argument.** Try `burroughs run calc.wasm add i32:2 i32:40`.

Next: [run a real Go program](02-wasip1-go-program.md).
