package exoip

import (
	"fmt"
	"os"
)

func AssertSuccess(err error) {
	if err == nil {
		return
	}
	fmt.Println("fatal error:", err)
	os.Exit(1)
}
