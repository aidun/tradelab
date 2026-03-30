package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

var totalCoveragePattern = regexp.MustCompile(`total:\s+\(statements\)\s+([0-9.]+)%`)

func main() {
	var (
		filePath = flag.String("file", "coverage.out", "path to the Go coverage profile")
		minimum  = flag.Float64("min", 40, "minimum total statement coverage percentage")
	)

	flag.Parse()

	output, err := exec.Command("go", "tool", "cover", "-func="+*filePath).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "run go tool cover: %v\n%s", err, string(output))
		os.Exit(1)
	}

	total, err := parseTotalCoverage(string(output))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse total coverage: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Backend total coverage: %.1f%% (minimum %.1f%%)\n", total, *minimum)
	if total < *minimum {
		fmt.Fprintf(os.Stderr, "backend coverage gate failed: %.1f%% is below %.1f%%\n", total, *minimum)
		os.Exit(1)
	}
}

func parseTotalCoverage(profile string) (float64, error) {
	// The raw coverage profile does not contain the total. Reuse the standard
	// summary format by asking callers to provide `go tool cover -func` output
	// when needed, but also support plain profile data by failing clearly.
	matches := totalCoveragePattern.FindStringSubmatch(profile)
	if len(matches) != 2 {
		return 0, fmt.Errorf("total coverage line not found")
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}
