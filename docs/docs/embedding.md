---
wip: true
---

# Embedding guide: the Go API

*Written against `v0.12.0` from the package documentation (`go doc -all
github.com/scttfrdmn/burroughs`) and the code behind it. Behaviour marked **observed** was checked by running programs:
most of it is in the example programs this site's CI runs against every release (see the tutorials), and the rest
was checked with one-off programs during the site's fact pass at `v0.11.0`. A few doc comments in the package are out of date
([#1130](https://github.com/scttfrdmn/burroughs/issues/1130)). Where that matters it is noted below, and what is
described here is what the code does. The Go declarations on this page are excerpts of the API, labelled as
fragments; the runnable programs are in the tutorials.*

``` { .go .fragment }
import "github.com/scttfrdmn/burroughs"
```

The package has four entry points, one per kind of artifact:

| You have | Call | You get |
| --- | --- | --- |
| a core module whose exports you want to call | `burroughs.Instantiate` / `Config.Instantiate` | `*Instance` |
| a WASI preview 1 command (`_start`) | `WASIP1Config.Run` | the guest's exit code |
| a component whose typed exports you want to call | `burroughs.LoadComponent` / `ComponentConfig.LoadComponent` | `*Component` |
| a component exporting `wasi:cli/run` | `ComponentConfig.Run` | the guest's exit code |

Three detection helpers read a module's header and sections **without instantiating it**:

- `IsComponent(wasm) (bool, error)` reads the preamble's layer field.
- `IsWASIP1Command(wasm) (bool, error)` checks that the module imports `wasi_snapshot_preview1` and exports `_start`.
  It decodes with the default feature set.
- `WASIP1Config.IsCommand(wasm) (bool, error)` asks the same question, but decodes under that config's `Features`.

For all three, `(false, err)` means the bytes could not be read. `(false, nil)` means they are not that kind of
artifact. **Check the error.** A `wasip1` command whose decode needs a gated feature comes back `(false, err)` with
`ErrGated` inside, which does not mean "not a command".

---

## Core modules

### `Instantiate`, `Config`, `Instance`

``` { .go .fragment }
func Instantiate(wasm []byte) (*Instance, error)          // Config{}.Instantiate
func (c Config) Instantiate(wasm []byte) (*Instance, error)

type Config struct {
	Strict bool // a decline becomes ErrDeclined instead of running
}
```

`Instantiate` decodes, validates and instantiates. When it fails, the error matches exactly one classification, in
this order:

1. **`ErrGated`**: the decoder hit a construct from a proposal whose gate is off, for example `gc: feature gate
   disabled`. This is checked before malformed, because a gated module is well-formed.
2. **`ErrMalformed`**: the bytes are not a module this decoder can read. The wrapped text is the spec's own, for
   example `magic header not detected`.
3. **`ErrInvalid`**: the module is well-formed but breaks a typing rule. The wrapped text names the rule, for example
   `type mismatch: ...`.
4. **`ErrDeclined`**, under `Strict` only: the validator has no rule yet for some construct. Without `Strict` the
   module runs, and `Instance.Decline` reports what went unchecked.
5. **`ErrUnlinkable`**: the module imports something. **The core path supplies no imports at all**, so any import
   (function, memory, table, global) is refused when the module loads (decision 0082). The error wraps the engine's
   link failure: `unknown import: "env" "log"`.
6. **`*Trap`**: the module's start function trapped while it was being instantiated.

**`Config` has no feature set and no import set.** A core module cannot be given host functions, and cannot turn on
a default-off proposal. `Config` exists to carry `Strict`, which its doc comment says will become a no-op once the
validator is complete.

``` { .go .fragment }
func (in *Instance) Exports() []string
func (in *Instance) Call(name string, args ...Value) ([]Value, error)
func (in *Instance) Global(name string) (Value, error)
func (in *Instance) Decline() error
func (in *Instance) Deferred() error
```

- **`Exports`** lists the exported **functions** in declaration order. These are exactly the names `Call`
  accepts. Globals, memories and tables are not listed, and there is no way to list exported globals.
- **`Call`** checks the arguments' count and types against the signature before running anything, then returns the
  results in declaration order (multi-value works). **Observed** errors:
  - `interp: "add" takes 2 arguments, got 1`
  - `interp: "add" parameter 0 is i32, got i64`
  - `interp: no exported function "nope"`

  These caller mistakes match **none** of the sentinels: they are plain errors. A trap returns a `*Trap`. An
  instruction the interpreter does not implement returns `ErrUnsupported`, when it is reached rather than when the
  module loads.
- **A core instance stays callable after a trap** (observed), as the core spec requires. Components behave
  differently: see [`ErrInstanceLocked`](#errors).
- **`Global`** reads an exported global's current value. Asking for a name that is a function, or not exported, is
  an error (`interp: no exported global "add"`). There is no public way to write a global.
- **`Decline`** is `nil` when every construct was validated. When it is not, the message names what was not, for
  example `instruction not in this slice: ...`. The repository's tests note that **no module in the default feature
  set is known to produce a decline at the moment**, so in practice this is `nil`. Check it anyway, because the
  validator is still being filled in.
- **`Deferred`** reports an allocation shortfall during instantiation that could not be reported as a trap: a
  *defined* memory or table that failed to allocate. A nil error from `Instantiate` says the module did not die
  coming to life. A nil `Deferred` says it came to life completely.

**Concurrency is not documented** for `Instance`. Nothing in the package says it is safe to call one `Instance` from
several goroutines. Until it does, serialize calls yourself or use one instance per goroutine.

The `Call` signature takes **no `context.Context`**. That is deliberate (ADR 0085 amendment 1): it is released
surface and is not changing. A running core call cannot be cancelled from the API.

### `Value`, `Type`, `Kind`

`Value` is a tagged wasm value. You build one with a constructor and read it with an accessor:

| Construct | Read | Notes |
| --- | --- | --- |
| `I32(int32)` | `Int32()` | |
| `I64(int64)` | `Int64()` | |
| `F32(float32)` | `Float32()` | stored by bit pattern, so NaN payloads survive |
| `F64(float64)` | `Float64()` | same |
| `V128(lo, hi uint64)` | `V128() (lo, hi)` | |
| `ExternRef(id uint32)` | `ExternID() (uint32, bool)` | non-null externref. 0 is a valid identity, so use `IsNull` to test for null. |
| `NullRef(t Type) (Value, bool)` | `IsNull()` | false if `t` is not a nullable reference type |
| `HostRef(t Type, id uint32) (Value, bool)` | `ExternID()` | a bare host reference at a given reference type |
| (result only) | `I31() (uint32, bool)`, `RefPayload() (RefPayload, bool)` | GC values. Their contents do not cross the boundary. |

An accessor called on a value of the wrong kind returns the zero value. It never reinterprets the bits. Use
`v.Type().Kind()` to find out what a result is.

`Type` is comparable with `==`. It is built with `NumberType(KindI32)`, `AbstractRefType(KindFuncRef, true)` or
`TypedRefType(idx, null)`. `Kind` is a closed enum (`KindI32` … `KindTypedRef`). The ten GC reference kinds are
reachable only with the GC gate on, which the public API cannot do.

`Value.String()` and `ParseValue(string)` are inverses, and use the CLI's spelling: `i32:42`, `f32:nan:0x200000`,
`extern:3`, `null:func`. `ParseValue` accepts both the signed and unsigned spelling of an integer (observed:
`ParseValue("i32:4294967295")` prints as `i32:-1`). Of the `ref:...` spellings, only a host reference (`ref:host:N:<type>`) parses back. The others print, but cannot be parsed.

### `Trap`

``` { .go .fragment }
type Trap struct{ Reason string }
```

`Reason` is the spec's trap text, word for word: `integer divide by zero`, `unreachable`, `out of bounds memory
access`. `Error()` adds a prefix: `trap: integer divide by zero`. Match with `errors.As(err, &trap)`. Component traps
are `*Trap` too.

---

## WASI preview 1

``` { .go .fragment }
type WASIP1Config struct {
	Args     []string  // argv; nil means {"program"}
	Env      []string  // "KEY=VALUE"; nil means none
	Stdin    io.Reader // nil means os.Stdin
	Stdout   io.Writer // nil means os.Stdout
	Stderr   io.Writer // nil means os.Stderr
	Preopens []Preopen // nil means no filesystem at all
	Features []Feature // capabilities the guest needs; see below
}

type Preopen struct {
	Host     string // host directory
	Guest    string // the name the guest sees, e.g. "/data"
	Writable bool   // false (the zero value) is read-only
}

func (c WASIP1Config) Run(wasm []byte) (exitCode int, err error)
func (c WASIP1Config) IsCommand(wasm []byte) (bool, error)
```

`Run` decodes, validates, links the module against the preview-1 host, and runs `_start`.

- When the guest calls **`proc_exit(n)`**, `Run` returns `(n, nil)`. A non-zero exit is not an error.
- **A load, validate or link failure, or a trap,** returns `(0, err)`. So a trap never looks like exit 0, and an
  exit never looks like a failure.
- **A module with no `_start`** returns `(0, err)`. Observed: `wasi: _start: interp: no exported function
  "_start"`. That error matches no sentinel. Use `IsCommand` first if you are routing.

The capability model (decision 0083):

- **No default grant.** With no `Preopens` the guest has no filesystem. In the Go guest from tutorial 2, an open
  then fails with `Bad file number` (EBADF).
- **Paths are confined to the grant's resolved root.** A `..`, an absolute path or a symlink pointing outside it
  cannot escape, for reads or writes. Writes are confined with `os.Root`.
- **Read-only by default.** A write under a read-only grant fails with `ENOTCAPABLE`, which a Go guest reports as
  `Capabilities insufficient` (observed).
- **`Writable: true`** makes a scratch grant. It enables `path_create_directory`, `path_unlink_file`,
  `path_remove_directory`, `fd_pwrite`, and `path_open`'s create and truncate modes. Twenty other
  preview-1 functions still answer `ENOSYS` (ADR 0091's amendment counts 21; the host refuses 20).
- **Env is empty unless you set it.** The host process's environment is not inherited.

**A guest can start its own threads** (since `v0.12.0`). If the module imports `"burroughs" "spawn"`, with type
`(i32) -> ()` or `(i32) -> (i32)`, and exports `mstart` of type `(i32) -> ()`, `Run` serves the import: each call
starts a new agent running `mstart(arg)` on its own OS thread, over the module's shared memory, and the
result-bearing form returns the new agent's thread id. A module that imports the door without a matching `mstart` is
refused at link. Contract §2 T-1 names the door, and it is what the Phase 4 Go fork's runtime calls. A Go guest on
that fork must not call `runtime.LockOSThread` and then need another M, which crashes the fork's runtime
([#1157](https://github.com/scttfrdmn/burroughs/issues/1157)).

Note the stream defaults: **nil streams are the process's own** here, which is the opposite of `ComponentConfig`.

### `Feature`

``` { .go .fragment }
type Feature string
const FeatureThreads Feature = "threads"
```

`Feature` is how you tell the engine what your artifact needs. It is accepted by `WASIP1Config` and
`ComponentConfig`, and not by `Config`. The only defined value is `FeatureThreads`. **An unrecognized name is
refused as `ErrUnsupported`** (`unknown feature "gc"; this build recognizes "threads"`). Supplying a feature does
not flip any gate's default.

`FeatureThreads` is redundant while `gate:threads` is on by default, which it has been since ADR 0106 (released in `v0.10.0`; the claim covers linux and darwin).
It matters when `BURROUGHS_THREADS=0` rolls that back: it admits shared memories and atomics again for that guest
(observed). The doc comment on `FeatureThreads` still says *"`gate:threads`' default is unchanged"*. That was true
when it was written and is not true now.

**There is no `Feature` for GC or exception handling.** Modules using them are `ErrGated` through every public path.
Extended constant expressions, tail calls, multi-memory and memory64 need no `Feature`: they are on by default since
`v0.12.0`, and only their environment rollbacks (`BURROUGHS_<GATE>=0`) turn them off.

---

## Components

### Loading

``` { .go .fragment }
func LoadComponent(wasm []byte) (*Component, error)                     // ComponentConfig{}.LoadComponent
func (c ComponentConfig) LoadComponent(wasm []byte) (*Component, error)
func (c ComponentConfig) Run(wasm []byte) (exitCode int, err error)

type ComponentConfig struct {
	Args     []string  // argv, via wasi:cli/environment.get-arguments
	Stdin    io.Reader // nil means an EMPTY stream (not os.Stdin)
	Stdout   io.Writer // nil DISCARDS (not os.Stdout)
	Stderr   io.Writer // nil DISCARDS (not os.Stderr)
	Features []Feature
}
```

**The nil defaults are the opposite of `WASIP1Config`'s on purpose.** A component is embedded in your program, so
it gets no terminal unless you hand it a writer. There is no `Env` and no `Preopens`: a component gets no environment
variables and no filesystem through this type.

At load:

- a component using a gated proposal (or any component, with `BURROUGHS_COMPONENTS=0`) returns `ErrGated`;
- a form the engine does not model returns `ErrUnsupported`.

`ComponentConfig.Run` calls the `wasi:cli/run` export and returns `(code, nil)` for a guest exit, or `(0, err)` for a
load or link failure or a trap. That is the same one-channel rule `WASIP1Config.Run` follows.

The `ComponentConfig` doc comment still says the value-call surface *"has not landed"*. It has: `LoadComponent`,
`Component.Call` and `ComponentValue` are all public in `v0.12.0`.

### `Component`

``` { .go .fragment }
func (c *Component) Exports() []string
func (c *Component) Call(ctx context.Context, name string, args ...ComponentValue) ([]ComponentValue, error)
func (c *Component) Close() error
```

- **Names**: a top-level function is called by its bare name (`"addrec"`). A function inside an exported
  interface is called as `"interface#function"`, for example `"wasi:cli/run@0.3.0#run"`. A bare name is never
  resolved by searching interfaces. `Exports()` returns every callable name, sorted, in exactly the form `Call`
  accepts. Interface names themselves are not in the list, because they are not callable.
- **Concurrent calls**: a `Component` may be called from several goroutines at once, and those calls are all in
  flight together, **but they execute one at a time**. Another call can run only when the current one suspends
  (ADR 0101). For parallel guest execution, use separate `Component`s.
- **Cancellation**: cancelling `ctx` delivers the component model's cancel event. The guest runs its own
  cancellation path, and `Call` returns `ErrCancelled` once the task has actually ended. A call that finishes before
  the cancellation takes effect returns its result normally. *(Not demonstrated in these docs: no fixture used here
  suspends long enough to be cancelled.)*
- **Traps lock the instance**, as they do in the component-model reference and in wasmtime. The trapping call
  returns its `*Trap`. Every later call returns `ErrInstanceLocked`, which wraps the same `*Trap` (observed). A call
  already in flight on another goroutine ends with `ErrInstanceLocked` at its next suspension point.
- **`Close`** cancels in-flight calls, waits a bounded time, then releases the component's engine threads and host
  state. It returns `ErrCloseIncomplete` if the bound was reached. It is idempotent: a second `Close` returns `nil`
  (observed). After `Close`, `Call` returns `ErrComponentClosed` (observed: `the component is closed: "ok" was not
  called`). Close every component you load.

### `ComponentValue` and `ComponentType`

Four WIT kinds cross the boundary in `v0.12.0`, in both directions:

| WIT | Type | Construct | Read |
| --- | --- | --- | --- |
| `u32` | `ComponentTypeU32()` | `ComponentU32(uint32)` (cannot fail) | `U32() (uint32, bool)` |
| `string` | `ComponentTypeString()` | `ComponentString(string) (ComponentValue, error)`: refuses invalid UTF-8 | `Str() (string, bool)` |
| `list<T>` | `ComponentTypeList(elem) (ComponentType, error)` | `ComponentList(elem, vals...) (ComponentValue, error)` | `List() ([]ComponentValue, bool)` |
| `record` | `ComponentTypeRecord(fields...) (ComponentType, error)` | `ComponentRecord(t, map[string]ComponentValue) (ComponentValue, error)` | `Field(name) (ComponentValue, bool)` |

A WIT `tuple` comes back as a record, because that is what the Canonical ABI turns it into.

The rules behind this table:

- **Mistakes are refused at construction.** `ComponentList` checks every element's type structurally (observed:
  `list element 0 is string, but the list's element type is u32`). `ComponentRecord` needs exactly the declared
  keys, so a missing field (`record field "b" is missing`) or an extra one is refused. `ComponentTypeRecord`
  refuses an empty field set, a duplicate label, or a label that is not valid WIT.
- **Field order comes from the type**, not from the map. Pass fields to `ComponentTypeRecord` in the order the
  component's WIT declares them, because that order is the byte layout.
- **A second check runs at the call.** A well-formed value passed to a parameter of a different type is refused as
  **the caller's error**, matching no sentinel, as a mistyped `Instance.Call` argument is. The message names the
  parameter, the declared type and the given type (observed, in [tutorial 3](tutorials/03-components.md)'s output).
  Until `v0.12.0` it was classed `ErrUnsupported`, which wrongly blamed the engine
  ([#1129](https://github.com/scttfrdmn/burroughs/issues/1129), fixed).
- **`Str()` versus `String()`**: `Str()` returns a string value's *content*. `String()` is a debug rendering
  (`string("burroughs")`, `u32(42)`) and is truncated for long values.
- **The zero `ComponentValue` and the zero `ComponentType` name no type.** They are refused wherever they are used.
- **Every other kind is refused by name, as `ErrUnsupported`**: `bool`, the other integer widths, floats, `char`,
  `variant`, `enum`, `option`, `result`, `flags`, and `own`/`borrow` resource handles. The `ComponentKind` enum lists
  them so that a refusal can name what it refused. (`ComponentKind`'s and `ComponentU32`'s doc comments still say
  only `u32` crosses. That is out of date.)
- There is no accessor for a returned value's type beyond `Kind()`. You are expected to know the signature you
  called (ADR 0097).

---

## Errors

Every sentinel is matched with `errors.Is`. The wrapped message carries the detail.

| Sentinel | Means | Where |
| --- | --- | --- |
| `ErrMalformed` | not a module this decoder can read | load |
| `ErrInvalid` | well-formed, breaks a typing rule | load |
| `ErrDeclined` | validator has no rule yet for a construct | load, `Strict` only (`Instance.Decline` otherwise) |
| `ErrGated` | well-formed, uses a proposal whose gate is off | load |
| `ErrUnlinkable` | an import nothing supplies | load (core path) |
| `ErrUnsupported` | the engine does not implement this, or an unknown `Feature`, or a component kind or form not modelled | load or point of use |
| `ErrCancelled` | a `Component.Call` whose context was cancelled, after the task ended | component call |
| `ErrComponentClosed` | a `Component.Call` after `Close` | component call |
| `ErrInstanceLocked` | a `Component.Call` after an earlier trap. Wraps that `*Trap`. | component call |
| `ErrCloseIncomplete` | `Close` hit its bound with calls still cancelling | `Close` |
| `*Trap` (type) | the module ran correctly and the program trapped | call or run |

A sketch of a complete switch is in [tutorial 4](tutorials/04-traps-and-refusals.md).
