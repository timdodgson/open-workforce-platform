package main

import (
	"fmt"
	"os"
)

// warnDeprecated prints a one-line notice for legacy commands.
func warnDeprecated(command, alternative string) {
	fmt.Fprintf(os.Stderr, "Note: %s uses the legacy application layer. Prefer %s.\n", command, alternative)
}
