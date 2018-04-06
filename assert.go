package exoip

import (
	"fmt"
	"os"
)

// assertSuccess logs the error and exists if something bad occurs
func assertSuccess(err error) {
	if err == nil {
		return
	}
	Logger.Crit(fmt.Sprintf("fatal: %s", err))
	fmt.Fprintln(os.Stderr, "fatal error:", err)
	os.Exit(1)
}
