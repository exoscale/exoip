package exoip

import (
	"fmt"
	"os"
)

func AssertSuccess(err error) {
	if err == nil {
		return
	}
	Logger.Crit(fmt.Sprintf("fatal: %s", err))
	fmt.Fprintln(os.Stderr, "fatal error:", err)
	os.Exit(1)
}
