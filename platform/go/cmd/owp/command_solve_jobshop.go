package main

import "os"

func runSolveJobShop() {
	warnDeprecatedSolveAlias("solve-jobshop", "jobshop")
	runSolveDomain("jobshop", os.Args[2:])
}
