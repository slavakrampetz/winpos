package cmd

import (
	"winpos/internal/lib"
)

func Reset() {
	lib.Cleanup()
}
