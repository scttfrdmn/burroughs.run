---
wip: true
---

# Tutorial 2: run a Go program built for WASI preview 1

*The code is this site's `examples/wasip1/` (`guest/`, `host/`, `data/`, `output.txt`). CI builds the guest with the
stock Go toolchain (the site pins go1.27.3), builds the `burroughs` CLI at the latest release tag, runs both the CLI
sessions and the Go host, and fails if any output differs from what is shown here.*

The stock Go toolchain can target `GOOS=wasip1 GOARCH=wasm`. (`wasip2`/`wasip3` are not supported Go targets.) A
`wasip1` binary is a *command*: it imports `wasi_snapshot_preview1` and exports `_start`. Burroughs runs one through
`WASIP1Config.Run`, or `burroughs run` on the command line.

## 1. The guest

`guest/main.go` is an ordinary Go program:

```go title="examples/wasip1/guest/main.go"
--8<-- "examples/wasip1/guest/main.go"
```

Build it. The file it reads, `data/greeting.txt`, holds the line `hello from the host`:

``` { .console .fragment }
$ cd guest
$ GOOS=wasip1 GOARCH=wasm go build -o ../guest.wasm .
```

(CI runs exactly this build; the binary is about 2.6 MB, so the site does not commit it.)

## 2. Run it from the CLI

From `examples/wasip1/`, where `in.txt` holds three lines (`a`, `b`, `c`). `[exit N]` is the exit status when it is
not 0:

```console title="examples/cli/transcripts/t2-run.txt"
--8<-- "examples/cli/transcripts/t2-run.txt"
```

The guest's exit code is the CLI's:

```console title="examples/cli/transcripts/t2-fail.txt"
--8<-- "examples/cli/transcripts/t2-fail.txt"
```

With no grants the guest sees no filesystem at all:

```console title="examples/cli/transcripts/t2-nogrants.txt"
--8<-- "examples/cli/transcripts/t2-nogrants.txt"
```

That `3` is the guest's own `os.Exit(3)`. It is not the CLI's "refused" code, even though the number is the same.
The difference is that there is no `burroughs:` line on stderr.

## 3. Run it from Go

`host/main.go`:

```go title="examples/wasip1/host/main.go"
--8<-- "examples/wasip1/host/main.go"
```

Run it with `go run .` in `examples/wasip1/host/`. It prints:

```text title="examples/wasip1/output.txt"
--8<-- "examples/wasip1/output.txt"
```

## What to take from it

- **`proc_exit(n)` is `(n, nil)`**. An error from `Run` means the module failed to load, link or run (a trap), and
  then the code is 0.
- **Capabilities are explicit.** No `Preopens` means no filesystem. A grant is read-only unless `Writable` is set.
  A write through a read-only grant fails in the guest with `ENOTCAPABLE` ("Capabilities insufficient"). Paths cannot
  escape a grant's resolved root.
- **Nothing is inherited from the host's environment.** The guest saw exactly the one variable it was given.
- **Stream defaults**: a nil `Stdin`/`Stdout`/`Stderr` in `WASIP1Config` means the *process's* streams. The host
  above sets all three so it can capture them. Components default the other way: see
  [tutorial 3](03-components.md).
- **Threads**: a guest from the stock Go toolchain runs the same with `BURROUGHS_THREADS=0`, so it does not depend
  on the threads gate:

    ```console title="examples/cli/transcripts/t2-threads0.txt"
    --8<-- "examples/cli/transcripts/t2-threads0.txt"
    ```

    The guests from the project's Go fork (the "Phase 4 fork") do carry atomics. Since `v0.12.0` their
  `"burroughs" "spawn"` import is served by `WASIP1Config.Run` and `burroughs run`, so a fork-built program starts
  real threads through this same path. One fork-side limit applies: a guest that calls `runtime.LockOSThread` and
  then needs another M crashes the fork's runtime. See [Known limits](../limits-and-status.md).
- Only a subset of preview 1 is implemented for writes. Twenty functions still answer `ENOSYS` (ADR 0091's amendment counts 21; the host refuses 20).
  If your guest needs one, it gets that error back. It does not crash.
