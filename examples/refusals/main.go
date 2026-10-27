// Command refusals loads one module per outcome and prints which public
// classification each one lands in.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/burroughs"
)

// classify names the sentinel an error matches. The sentinels are disjoint
// classifications, so the order of the cases does not matter for these.
func classify(err error) string {
	var trap *burroughs.Trap
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, burroughs.ErrMalformed):
		return "ErrMalformed"
	case errors.Is(err, burroughs.ErrInvalid):
		return "ErrInvalid"
	case errors.Is(err, burroughs.ErrDeclined):
		return "ErrDeclined"
	case errors.Is(err, burroughs.ErrGated):
		return "ErrGated"
	case errors.Is(err, burroughs.ErrUnlinkable):
		return "ErrUnlinkable"
	case errors.Is(err, burroughs.ErrUnsupported):
		return "ErrUnsupported"
	case errors.As(err, &trap):
		return "*Trap"
	}
	return "other"
}

func main() {
	for _, name := range []string{"malformed", "invalid", "memory64", "tailcall", "imports", "shared"} {
		wasm, err := os.ReadFile(name + ".wasm")
		if err != nil {
			log.Fatal(err)
		}
		in, err := burroughs.Config{Strict: true}.Instantiate(wasm)
		fmt.Printf("%-9s -> %s\n", name, classify(err))
		if err != nil {
			fmt.Printf("            %v\n", err)
			continue
		}
		if name == "shared" {
			// Admitted by default (gate:threads is on), but nothing can start a
			// second agent over this memory, so a wait can only time out.
			n, err := in.Call("inc")
			fmt.Println("            inc ->", n, err)
			r, err := in.Call("wait")
			fmt.Println("            wait(timeout 1ms) ->", r, err, "(2 means timed-out)")
		}
	}
}
