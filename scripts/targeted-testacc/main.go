// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "targeted-testacc: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		base             string
		totalShards      int
		shardIndex       int
		dryRun           bool
		verbose          bool
		runAllThreshold  float64
		minShardPackages int
	)

	flag.StringVar(&base, "base", "", "git diff baseline (overrides TARGETED_TESTACC_BASE)")
	flag.IntVar(&totalShards, "total-shards", 1, "total number of shards")
	flag.IntVar(&shardIndex, "shard-index", 0, "index of this shard (0-based)")
	flag.BoolVar(&dryRun, "dry-run", false, "print selection rationale instead of package list")
	flag.BoolVar(&verbose, "verbose", false, "print additional diagnostics")
	flag.Float64Var(&runAllThreshold, "run-all-threshold", 70.0, "percentage of acc-test packages that triggers a full run")
	flag.IntVar(&minShardPackages, "min-shard-packages", 30, "minimum selected packages before multi-shard splitting is used")
	flag.Parse()

	if err := validateFlags(totalShards, shardIndex, runAllThreshold, minShardPackages); err != nil {
		return err
	}

	modulePath, err := currentModulePath()
	if err != nil {
		return fmt.Errorf("resolve module path: %w", err)
	}

	baseline := ResolveBaseline(base)
	if verbose {
		fmt.Fprintf(os.Stderr, "using diff baseline: %s\n", baseline)
	}

	changedFiles, err := GitDiff(baseline)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "git diff failed: %v\n", err)
		}
		changedFiles = nil
	}

	if dryRun {
		fmt.Println("Changed files:")
		if len(changedFiles) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, f := range changedFiles {
				fmt.Printf("  %s\n", f)
			}
		}
	}

	classifier := NewClassifier(modulePath)
	var classified *ClassifyResult
	if len(changedFiles) == 0 {
		classified = &ClassifyResult{ForceAll: true}
		if dryRun {
			fmt.Println("\nNo resolvable diff; selecting all acceptance test packages.")
		}
	} else {
		classified = classifier.Classify(changedFiles)

		if classified.ForceAll {
			if dryRun {
				fmt.Println("\nForce-all prefix matched; selecting all acceptance test packages.")
			}
		} else if !classified.HasCode {
			if dryRun {
				fmt.Println("\nNo changed Go or testdata files; zero packages selected.")
			}
			return nil
		}
	}

	allAccPackages, err := FindAccTestPackages("internal", modulePath)
	if err != nil {
		return fmt.Errorf("enumerate acceptance test packages: %w", err)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "found %d acceptance test packages\n", len(allAccPackages))
	}

	phase1Packages := []string{}
	phase2Packages := []string{}
	phaseReasons := make(map[string][]string)

	accSet := make(map[string]struct{}, len(allAccPackages))
	for _, p := range allAccPackages {
		accSet[p] = struct{}{}
	}

	if !classified.ForceAll {
		graph, err := BuildImportGraph()
		if err != nil {
			return fmt.Errorf("build import graph: %w", err)
		}

		// Phase 1: reverse dependency walk intersected with acc-test packages.
		transitive := WalkReverseDeps(graph.Reverse, classified.Packages)
		for _, p := range transitive {
			if _, ok := accSet[p]; ok {
				phase1Packages = append(phase1Packages, p)
				phaseReasons[p] = append(phaseReasons[p], "phase-1 reverse dependency")
			}
		}
		sort.Strings(phase1Packages)

		// Phase 2: entity grep across testdata and _test.go files.
		pkgDir := func(importPath string) string {
			return strings.TrimPrefix(importPath, modulePath+"/")
		}

		for _, pkg := range classified.Packages {
			entities, err := ExtractEntities(pkgDir(pkg))
			if err != nil {
				return fmt.Errorf("extract entities for %s: %w", pkg, err)
			}
			for _, ent := range entities {
				consumers, err := FindTestConsumers("internal", modulePath, ent.FullName())
				if err != nil {
					return fmt.Errorf("find consumers for %s: %w", ent.FullName(), err)
				}
				for _, consumer := range consumers {
					if _, ok := accSet[consumer]; !ok {
						continue
					}
					phaseReasons[consumer] = append(phaseReasons[consumer], fmt.Sprintf("phase-2 consumer of %s", ent.FullName()))
					phase2Packages = append(phase2Packages, consumer)
				}
			}
		}
		phase2Packages = stringsSorted(phase2Packages)
	}

	selected := SelectPackages(classified.ForceAll, phase1Packages, phase2Packages, allAccPackages, runAllThreshold)
	if verbose {
		fmt.Fprintf(os.Stderr, "selected %d packages before sharding\n", len(selected))
	}

	sharded := ApplyShard(selected, totalShards, shardIndex, minShardPackages)

	if dryRun {
		printDryRun(selected, phaseReasons, sharded, totalShards, shardIndex, minShardPackages)
		return nil
	}

	for _, pkg := range sharded {
		fmt.Println(pkg)
	}
	return nil
}

func validateFlags(totalShards, shardIndex int, runAllThreshold float64, minShardPackages int) error {
	if totalShards <= 0 {
		return fmt.Errorf("--total-shards must be >= 1")
	}
	if shardIndex < 0 {
		return fmt.Errorf("--shard-index must be >= 0")
	}
	if runAllThreshold < 0 || runAllThreshold > 100 {
		return fmt.Errorf("--run-all-threshold must be between 0 and 100")
	}
	if minShardPackages < 0 {
		return fmt.Errorf("--min-shard-packages must be >= 0")
	}
	return nil
}

func currentModulePath() (string, error) {
	out, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return "", fmt.Errorf("go list -m: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func printDryRun(selected []string, reasons map[string][]string, sharded []string, totalShards, shardIndex, minShardPackages int) {
	fmt.Printf("\nFull selected package set (%d packages):\n", len(selected))
	for _, pkg := range selected {
		fmt.Printf("  %s\n", pkg)
		if rs := reasons[pkg]; len(rs) > 0 && len(selected) <= 200 {
			for _, r := range rs {
				fmt.Printf("    - %s\n", r)
			}
		}
	}

	fmt.Printf("\nShard assignment (total-shards=%d shard-index=%d min-shard-packages=%d):\n", totalShards, shardIndex, minShardPackages)
	if len(sharded) == 0 {
		fmt.Println("  (no packages for this shard)")
	} else {
		for _, pkg := range sharded {
			fmt.Printf("  %s\n", pkg)
		}
	}
	fmt.Printf("\n%s\n", dryRunSummary(len(selected), len(sharded)))
}

func dryRunSummary(selected, sharded int) string {
	if selected == 0 {
		return "Result: 0 packages selected."
	}
	if sharded == 0 {
		return fmt.Sprintf("Result: %d packages selected, 0 emitted for this shard.", selected)
	}
	return fmt.Sprintf("Result: %d packages selected, %d emitted for this shard.", selected, sharded)
}
