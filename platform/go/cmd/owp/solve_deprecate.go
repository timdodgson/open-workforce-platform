package main

import (
	"fmt"
	"os"
)

func warnDeprecatedSolveAlias(legacyCmd, domain string) {
	fmt.Fprintf(os.Stderr, "Deprecated: %q — use: owp solve %s ...\n", legacyCmd, domain)
}
