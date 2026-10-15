package data

import (
	"os"
)

// IsStdinPiped returns true if stdin has piped data.
func IsStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice == 0
}
