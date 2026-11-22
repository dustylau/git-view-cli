// slowgit lets the terminal smoke test interrupt an in-flight Git read.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	// The application must terminate this subprocess even if it ignores Ctrl+C.
	signal.Ignore(os.Interrupt)
	ready := os.Getenv("GIT_VIEW_TEST_READY")
	if ready == "" {
		os.Exit(2)
	}
	if err := os.WriteFile(ready, fmt.Appendf(nil, "%d\n", os.Getpid()), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	time.Sleep(15 * time.Second)
	os.Exit(1)
}
