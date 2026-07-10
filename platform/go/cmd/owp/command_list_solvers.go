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
			line := fmt.Sprintf("  %s", name)
			if desc.Command != "" {
				line += fmt.Sprintf(" (%s)", desc.Command)
			}
			if desc.Usage != "" {
				line += " — " + desc.Usage
			}
			fmt.Println(line)
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
