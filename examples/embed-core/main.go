package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/burroughs"
)

func main() {
	wasm, err := os.ReadFile("calc.wasm")
	if err != nil {
		log.Fatal(err)
	}

	in, err := burroughs.Instantiate(wasm)
	if err != nil {
		log.Fatal(err) // ErrMalformed, ErrInvalid, ErrGated, ErrUnlinkable, ...
	}
	if d := in.Decline(); d != nil {
		fmt.Println("ran with a construct the validator could not check:", d)
	}
	fmt.Println("exports:", in.Exports())

	sum, err := in.Call("add", burroughs.I32(2), burroughs.I32(40))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("add:", sum[0].Int32(), "spelled", sum[0])

	prod, err := in.Call("mul64", burroughs.I64(1<<20), burroughs.I64(1<<20))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("mul64:", prod[0].Int64())

	h, err := in.Call("hypot", burroughs.F64(3), burroughs.F64(4))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("hypot:", h[0].Float64(), "type", h[0].Type())

	// Multiple results come back in declaration order.
	qr, err := in.Call("divmod", burroughs.I32(17), burroughs.I32(5))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("divmod:", qr[0], qr[1])

	// Arguments are checked against the signature before anything runs.
	if _, err := in.Call("add", burroughs.I64(1), burroughs.I32(2)); err != nil {
		fmt.Println("wrong argument type:", err)
	}
	if _, err := in.Call("add", burroughs.I32(1)); err != nil {
		fmt.Println("wrong arity:", err)
	}
	if _, err := in.Call("nope"); err != nil {
		fmt.Println("no such export:", err)
	}

	// A trap is a *burroughs.Trap, with the spec's text in Reason.
	var trap *burroughs.Trap
	if _, err := in.Call("divmod", burroughs.I32(1), burroughs.I32(0)); errors.As(err, &trap) {
		fmt.Printf("trap: %q (Error(): %v)\n", trap.Reason, err)
	}
	if _, err := in.Call("boom"); errors.As(err, &trap) {
		fmt.Printf("trap: %q\n", trap.Reason)
	}

	// A core instance stays callable after a trap.
	again, err := in.Call("add", burroughs.I32(1), burroughs.I32(1))
	fmt.Println("after the traps:", again, err)

	// Exported globals are read by name; there is no public write.
	calls, err := in.Global("calls")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("global calls =", calls)
	if _, err := in.Global("add"); err != nil {
		fmt.Println("reading a function as a global:", err)
	}

	// The CLI's value spelling round-trips through ParseValue.
	v, err := burroughs.ParseValue("i32:4294967295")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ParseValue(\"i32:4294967295\") =", v, v.Int32())
}
