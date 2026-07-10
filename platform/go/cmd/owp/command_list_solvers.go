package main

import (
	"fmt"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

func runListSolvers() {
	fmt.Println("Registered problems:")
	names := sdk.Problems()
	if len(names) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, name := range names {
			desc, _ := sdk.GetProblem(name)
			fmt.Printf("  %s — owp solve %s --instance <path>", name, name)
			if desc.Usage != "" {
				fmt.Printf("  (%s)", trimSolveUsage(desc.Usage))
			}
			fmt.Println()
		}
	}
	fmt.Println()
	fmt.Println("Built-in search modes:", strings.Join(sdk.BuiltInSearchModes, ", "))
	custom := sdk.RegisteredSearchModes()
	fmt.Print("Custom search modes: ")
	if len(custom) == 0 {
		fmt.Println("(none)")
	} else {
		fmt.Println(strings.Join(custom, ", "))
	}
}

func trimSolveUsage(usage string) string {
	const prefix = "owp solve "
	if strings.HasPrefix(usage, prefix) {
		rest := strings.TrimPrefix(usage, prefix)
		if idx := strings.Index(rest, " --instance"); idx > 0 {
			return rest[:idx]
		}
	}
	return usage
}
