package main

import (
	"fmt"
	"os"

	"github.com/RokyErickson/doppelganger/pkg/daemon"
)

func main() {
	if lock, err := daemon.AcquireLock(); err != nil {
		fmt.Fprintln(os.Stderr, "Doppelganger lock acquisition failed")
		os.Exit(1)
	} else {
		lock.Unlock()
	}
}
