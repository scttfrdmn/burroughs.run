// Command guest is an ordinary Go program, built for GOOS=wasip1 GOARCH=wasm.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("args:", os.Args)
	fmt.Println("GREETING env:", os.Getenv("GREETING"))
	fmt.Println("env entries:", len(os.Environ()))

	b, err := os.ReadFile("/data/greeting.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(3)
	}
	fmt.Printf("/data/greeting.txt: %q\n", strings.TrimSpace(string(b)))

	// Writing under a read-only grant is refused.
	if err := os.WriteFile("/data/new.txt", []byte("x"), 0o644); err != nil {
		fmt.Println("write under /data:", err)
	}

	// Writing under a writable (scratch) grant, if one was given.
	if err := os.WriteFile("/out/result.txt", []byte("written by the guest\n"), 0o644); err != nil {
		fmt.Println("write under /out:", err)
	} else {
		fmt.Println("wrote /out/result.txt")
	}

	// Read stdin lines, if any.
	sc := bufio.NewScanner(os.Stdin)
	n := 0
	for sc.Scan() {
		n++
	}
	fmt.Println("stdin lines:", n)

	if len(os.Args) > 1 && os.Args[1] == "fail" {
		os.Exit(7)
	}
}
