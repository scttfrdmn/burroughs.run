// Command describe runs tutorial 4's branching template over one error of
// each kind it can produce without fixtures: success, a malformed module, a
// gated module, a trap, an unknown Feature, and a caller mistake.
package main

import (
	"errors"
	"fmt"

	"github.com/scttfrdmn/burroughs"
)

// --8<-- [start:describe]
func describe(err error) string {
	var trap *burroughs.Trap
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, burroughs.ErrInstanceLocked): // component: an earlier trap locked it (also wraps a *Trap)
		return "locked"
	case errors.As(err, &trap): // ran, then trapped: trap.Reason
		return "trap: " + trap.Reason
	case errors.Is(err, burroughs.ErrGated): // fine module, gate off
		return "gated"
	case errors.Is(err, burroughs.ErrMalformed),
		errors.Is(err, burroughs.ErrInvalid),
		errors.Is(err, burroughs.ErrDeclined),
		errors.Is(err, burroughs.ErrUnlinkable): // the module was refused
		return "refused"
	case errors.Is(err, burroughs.ErrUnsupported): // engine gap, unknown Feature, or (until v0.12.0, #1129) a component type mismatch at Call
		return "unsupported"
	case errors.Is(err, burroughs.ErrCancelled): // you cancelled ctx, and the task ended
		return "cancelled"
	case errors.Is(err, burroughs.ErrComponentClosed):
		return "closed"
	case errors.Is(err, burroughs.ErrCloseIncomplete):
		return "close incomplete"
	default: // a caller mistake: wrong arity, no such export, ...
		return "other: " + err.Error()
	}
}

// --8<-- [end:describe]

// boom is a module exporting one function, "boom", whose body is unreachable.
var boom = []byte{
	0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // magic, version 1
	0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // type section: () -> ()
	0x03, 0x02, 0x01, 0x00, // function section: one function of type 0
	0x07, 0x08, 0x01, 0x04, 'b', 'o', 'o', 'm', 0x00, 0x00, // export "boom" = func 0
	0x0a, 0x05, 0x01, 0x03, 0x00, 0x00, 0x0b, // code: no locals, unreachable, end
}

// gcType is (module (type (struct (field i32)))), which needs the GC gate: off by default. (This was a memory64
// module until v0.12.0 turned memory64 on.)
var gcType = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x05, 0x01, 0x5f, 0x01, 0x7f, 0x00}

func main() {
	in, err := burroughs.Instantiate(boom)
	fmt.Println("instantiate boom:", describe(err))

	_, err = in.Call("boom")
	fmt.Println("call boom:       ", describe(err))

	_, err = in.Call("nope")
	fmt.Println("call nope:       ", describe(err))

	_, err = burroughs.Instantiate([]byte("not a wasm module"))
	fmt.Println("malformed bytes: ", describe(err))

	_, err = burroughs.Instantiate(gcType)
	fmt.Println("GC module:       ", describe(err))

	_, err = burroughs.WASIP1Config{Features: []burroughs.Feature{"gc"}}.Run(boom)
	fmt.Println("Feature \"gc\":    ", describe(err))
}
