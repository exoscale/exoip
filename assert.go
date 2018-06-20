package exoip

import (
	"fmt"
	"os"
)

// assertSuccessOrExit logs the error and exists if something bad occurs
func assertSuccessOrExit(err error) {
	if err == nil {
		return
	}
	Logger.Crit(fmt.Sprintf("fatal: %s", err))
	_, err = fmt.Fprintf(os.Stderr, "fatal error: %s\n", err)
	if err != nil {
		panic(err)
	}
	os.Exit(1)
}
