// Command component loads four of the Burroughs repository's committed component
// fixtures and calls their typed exports, then runs a wasi:cli/run component.
// The fixtures are fetched into testdata/ from the engine repository at the
// release tag (internal/component/testdata/); scripts/run-examples.sh does it.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/burroughs"
)

const fixtures = "testdata/"

func load(name string) *burroughs.Component {
	wasm, err := os.ReadFile(fixtures + name)
	if err != nil {
		log.Fatal(err)
	}
	if ok, err := burroughs.IsComponent(wasm); !ok {
		log.Fatalf("%s is not a component: %v", name, err)
	}
	c, err := burroughs.LoadComponent(wasm)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func main() {
	ctx := context.Background()

	// 1. A record argument, a u32 result.
	rc := load("record-synth.wasm")
	defer rc.Close()
	fmt.Println("record-synth exports:", rc.Exports())
	pair, err := burroughs.ComponentTypeRecord(
		burroughs.ComponentField{Name: "a", Type: burroughs.ComponentTypeU32()},
		burroughs.ComponentField{Name: "b", Type: burroughs.ComponentTypeU32()},
	)
	if err != nil {
		log.Fatal(err)
	}
	rec, err := burroughs.ComponentRecord(pair, map[string]burroughs.ComponentValue{
		"a": burroughs.ComponentU32(10),
		"b": burroughs.ComponentU32(32),
	})
	if err != nil {
		log.Fatal(err)
	}
	res, err := rc.Call(ctx, "addrec", rec)
	if err != nil {
		log.Fatal(err)
	}
	sum, ok := res[0].U32()
	fmt.Println("addrec({a:10, b:32}) =", sum, ok, "rendered", res[0])

	// A record result: read its fields by name.
	res, err = rc.Call(ctx, "makerec")
	if err != nil {
		log.Fatal(err)
	}
	name, _ := res[0].Field("name")
	nfield, _ := res[0].Field("n")
	text, _ := name.Str()
	nval, _ := nfield.U32()
	fmt.Printf("makerec() = {name: %q, n: %d}  (String() of the name field: %v)\n", text, nval, name)

	// Construction-time refusals.
	_, err = burroughs.ComponentRecord(pair, map[string]burroughs.ComponentValue{"a": burroughs.ComponentU32(1)})
	fmt.Println("missing field:", err)
	_, err = burroughs.ComponentString("\xff")
	fmt.Println("invalid UTF-8:", err)
	_, err = rc.Call(ctx, "addrec", burroughs.ComponentU32(1))
	fmt.Println("wrong argument type at the call:", err)

	// 2. A string argument.
	sc := load("string-echo.wasm")
	defer sc.Close()
	fmt.Println("string-echo exports:", sc.Exports())
	s, err := burroughs.ComponentString("héllo, wörld")
	if err != nil {
		log.Fatal(err)
	}
	res, err = sc.Call(ctx, "run-sync", s)
	if err != nil {
		log.Fatal(err)
	}
	n, _ := res[0].U32()
	fmt.Println("run-sync(\"héllo, wörld\") =", n, "(bytes of UTF-8 the guest saw echoed)")

	// 2b. A list<u32> argument.
	lc := load("string-arg-realloc-synth.wasm")
	defer lc.Close()
	fmt.Println("string-arg-realloc-synth exports:", lc.Exports())
	xs, err := burroughs.ComponentList(burroughs.ComponentTypeU32(),
		burroughs.ComponentU32(1), burroughs.ComponentU32(2), burroughs.ComponentU32(39))
	if err != nil {
		log.Fatal(err)
	}
	res, err = lc.Call(ctx, "sum-list", xs)
	if err != nil {
		log.Fatal(err)
	}
	total, _ := res[0].U32()
	fmt.Println("sum-list([1, 2, 39]) =", total)
	_, err = burroughs.ComponentList(burroughs.ComponentTypeU32(), s)
	fmt.Println("a string in a list<u32>:", err)

	// 3. A trap locks the instance.
	tc := load("trap-then-call-plain.wasm")
	fmt.Println("trap-then-call-plain exports:", tc.Exports())
	_, err = tc.Call(ctx, "run")
	var trap *burroughs.Trap
	fmt.Println("run:", err, "| is *Trap:", errors.As(err, &trap))
	_, err = tc.Call(ctx, "ok")
	fmt.Println("ok after the trap:", err)
	fmt.Println("  errors.Is ErrInstanceLocked:", errors.Is(err, burroughs.ErrInstanceLocked),
		"| wraps the *Trap:", errors.As(err, &trap))
	fmt.Println("Close:", tc.Close(), "| second Close:", tc.Close())
	_, err = tc.Call(ctx, "ok")
	fmt.Println("call after Close:", err, errors.Is(err, burroughs.ErrComponentClosed))

	// 4. A wasi:cli/run component, with its stdout captured.
	wasm, err := os.ReadFile(fixtures + "p3hello.wasm")
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	code, err := burroughs.ComponentConfig{Stdout: &out}.Run(wasm)
	fmt.Printf("p3hello: exit %d, err %v, stdout %q\n", code, err, out.String())

	// The zero configuration discards output.
	code, err = burroughs.ComponentConfig{}.Run(wasm)
	fmt.Printf("p3hello with zero config: exit %d, err %v (stdout discarded)\n", code, err)
}
