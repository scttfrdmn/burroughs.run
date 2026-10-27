// Command host runs ../guest.wasm, a Go program built for wasip1, through
// burroughs.WASIP1Config, with captured stdout, an environment, a read-only
// directory and a writable scratch directory.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/scttfrdmn/burroughs"
)

func main() {
	wasm, err := os.ReadFile("../guest.wasm")
	if err != nil {
		log.Fatal(err)
	}

	isCmd, err := burroughs.IsWASIP1Command(wasm)
	fmt.Println("IsWASIP1Command:", isCmd, err)

	scratch, err := os.MkdirTemp("", "wasip1-scratch-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(scratch)
	data, err := filepath.Abs("../data")
	if err != nil {
		log.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cfg := burroughs.WASIP1Config{
		Args:   []string{"guest", "fail"},
		Env:    []string{"GREETING=hi"},
		Stdin:  strings.NewReader("one\ntwo\n"),
		Stdout: &stdout,
		Stderr: &stderr,
		Preopens: []burroughs.Preopen{
			{Host: data, Guest: "/data"},                  // read-only (the zero value)
			{Host: scratch, Guest: "/out", Writable: true}, // a scratch grant
		},
	}
	code, err := cfg.Run(wasm)
	fmt.Printf("exit code %d, err %v\n", code, err)
	fmt.Print("--- guest stdout\n", stdout.String())
	fmt.Print("--- guest stderr\n", stderr.String())

	written, err := os.ReadFile(filepath.Join(scratch, "result.txt"))
	fmt.Printf("--- host sees result.txt: %q %v\n", written, err)

	// No Preopens: the guest reaches no filesystem at all.
	stdout.Reset()
	stderr.Reset()
	code, err = burroughs.WASIP1Config{
		Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr,
	}.Run(wasm)
	fmt.Printf("--- no grants: exit code %d, err %v, stderr %q\n", code, err, stderr.String())
}
