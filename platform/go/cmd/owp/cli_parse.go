package main

import (
	"fmt"
	"os"
	"strings"
)

func parseIntFlag(args []string, flag string) int {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return atoiOrFail(args[i+1], flag)
		}
		if strings.HasPrefix(arg, flag+"=") {
			return atoiOrFail(strings.TrimPrefix(arg, flag+"="), flag)
		}
	}
	return 0
}

func parseFloatFlag(args []string, flag string) float64 {
	for i, arg := range args {
		var val string
		if arg == flag && i+1 < len(args) {
			val = args[i+1]
		} else if strings.HasPrefix(arg, flag+"=") {
			val = strings.TrimPrefix(arg, flag+"=")
		}
		if val != "" {
			return parseFloat(val, flag)
		}
	}
	return 0
}

func parseParallelFlag(args []string) bool {
	return parseBoolFlag(args, "--parallel") == "true"
}

func parseBoolFlag(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			v := strings.TrimSpace(args[i+1])
			if v != "true" && v != "false" {
				fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be true or false)\n", flag, v)
				os.Exit(1)
			}
			return v
		}
		if strings.HasPrefix(arg, flag+"=") {
			v := strings.TrimSpace(strings.TrimPrefix(arg, flag+"="))
			if v != "true" && v != "false" {
				fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be true or false)\n", flag, v)
				os.Exit(1)
			}
			return v
		}
	}
	return ""
}

func parseStringFlag(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, flag+"=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, flag+"="))
		}
	}
	return ""
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

func atoiOrFail(s, flag string) int {
	s = strings.TrimSpace(s)
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a positive integer)\n", flag, s)
			os.Exit(1)
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a positive integer)\n", flag, s)
		os.Exit(1)
	}
	return n
}

func parseFloat(s, flag string) float64 {
	s = strings.TrimSpace(s)
	var result float64
	var decimal float64
	inDecimal := false
	divisor := 1.0
	for _, ch := range s {
		if ch == '.' {
			if inDecimal {
				fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a number)\n", flag, s)
				os.Exit(1)
			}
			inDecimal = true
			continue
		}
		if ch < '0' || ch > '9' {
			fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a number)\n", flag, s)
			os.Exit(1)
		}
		if inDecimal {
			divisor *= 10
			decimal += float64(ch-'0') / divisor
		} else {
			result = result*10 + float64(ch-'0')
		}
	}
	result += decimal
	if result <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid %s value: %s (must be a positive number)\n", flag, s)
		os.Exit(1)
	}
	return result
}

func parseSeedList(s string) []int64 {
	parts := strings.Split(s, ",")
	var seeds []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n := int64(0)
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				fmt.Fprintf(os.Stderr, "Invalid seed value: %s (must be a positive integer)\n", p)
				os.Exit(1)
			}
			n = n*10 + int64(ch-'0')
		}
		if n <= 0 {
			fmt.Fprintf(os.Stderr, "Invalid seed value: %s (must be a positive integer)\n", p)
			os.Exit(1)
		}
		seeds = append(seeds, n)
	}
	if len(seeds) == 0 {
		fmt.Fprintln(os.Stderr, "No valid seeds provided")
		os.Exit(1)
	}
	return seeds
}

func parseShowInvalidFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--show-invalid" {
			return true
		}
	}
	return false
}

func requireInstanceFlag(args []string, usage string) string {
	instancePath := parseStringFlag(args, "--instance")
	if instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path> is required")
		if usage != "" {
			fmt.Fprintln(os.Stderr, usage)
		}
		os.Exit(1)
	}
	return instancePath
}
