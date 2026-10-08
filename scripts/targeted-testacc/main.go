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
	"encoding/json"
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
		verbose          bool
		runAllThreshold  float64
		minShardPackages int
	)

	flag.StringVar(&base, "base", "", "git diff baseline (overrides TARGETED_TESTACC_BASE)")
	flag.IntVar(&totalShards, "total-shards", 1, "total number of shards")
	flag.Bool("dry-run", false, "accepted for compatibility; the JSON shard plan is always emitted without running tests")
	flag.BoolVar(&verbose, "verbose", false, "print additional diagnostics")
	flag.Float64Var(&runAllThreshold, "run-all-threshold", 70.0, "percentage of acc-test packages that triggers a full run")
	flag.IntVar(&minShardPackages, "min-shard-packages", 30, "minimum selected packages before multi-shard splitting is used")
	flag.Parse()

	if err := validateFlags(totalShards, runAllThreshold, minShardPackages); err != nil {
		return err
	}

	modulePath, err := currentModulePath()
	if err != nil {
		return fmt.Errorf("resolve module path: %w", err)
	}

	baseline := ResolveBaseline(base)
	if verbose {
		fmt.Fprintf(os.Stderr, "using diff baseline: %s\n", baseline.Ref)
	}

	changedFiles, err := GitDiff(baseline)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: git diff against %s failed (%v); falling back to the full acceptance test suite\n", baseline.Ref, err)
		changedFiles = nil
	}

	rationale := selectionRationale(changedFiles)

	var classified *ClassifyResult
	if len(changedFiles) == 0 {
		classified = &ClassifyResult{ForceAll: true}
	} else {
		classified = NewClassifier(modulePath).Classify(changedFiles)
		if classified.ForceAll {
			rationale = append(rationale, "force-all prefix matched; selecting all acceptance test packages")
		} else if !classified.HasCode {
			rationale = append(rationale, "no changed files map to a Go package; zero packages selected")
		}
	}

	allAccPackages, err := FindAccTestPackages(accTestEnumerationRoots, modulePath)
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

	if !classified.ForceAll && classified.HasCode {
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
			if importPath == modulePath {
				return "."
			}
			return strings.TrimPrefix(importPath, modulePath+"/")
		}

		// Collect every entity name from every candidate package first, then
		// walk internal/ once for all of them, so the walk cost does not scale
		// with the number of entities.
		//
		// The candidate set is the union of the changed packages and the
		// packages phase 1 selects. Phase 1 walks reverse dependencies
		// transitively, including packages whose sub-packages declare entities;
		// running entity extraction over the changed set alone would miss
		// cross-package testdata consumers of those entities.
		//
		// Seed from the pre-filter transitive set rather than phase1Packages:
		// phase1Packages is already intersected with accSet, so an entity-
		// declaring package that transitively imports changed code but has no
		// acceptance tests of its own (e.g. a future internal/<newhelper>) would
		// otherwise be dropped before extraction and its cross-package .tf
		// consumers never found. ExtractEntities on such a package is a no-op
		// cost when it declares nothing.
		candidatePkgs := make(map[string]struct{})
		for _, pkg := range classified.Packages {
			candidatePkgs[pkg] = struct{}{}
		}
		for _, pkg := range transitive {
			candidatePkgs[pkg] = struct{}{}
		}

		entityNames := make([]string, 0)
		for pkg := range candidatePkgs {
			entities, err := ExtractEntities(pkgDir(pkg))
			if err != nil {
				return fmt.Errorf("extract entities for %s: %w", pkg, err)
			}
			for _, ent := range entities {
				entityNames = append(entityNames, ent.FullName())
			}
		}
		sort.Strings(entityNames)

		consumerPkgs, err := FindTestConsumersMulti(accTestEnumerationRoots, modulePath, entityNames)
		if err != nil {
			return fmt.Errorf("find test consumers: %w", err)
		}
		for consumer, matched := range consumerPkgs {
			if _, ok := accSet[consumer]; !ok {
				continue
			}
			phaseReasons[consumer] = append(phaseReasons[consumer], fmt.Sprintf("phase-2 consumer of %s", strings.Join(matched, ", ")))
			phase2Packages = append(phase2Packages, consumer)
		}
		phase2Packages = stringsSorted(phase2Packages)
	}

	selected := SelectPackages(classified.ForceAll, phase1Packages, phase2Packages, allAccPackages, runAllThreshold)
	if verbose {
		fmt.Fprintf(os.Stderr, "selected %d packages\n", len(selected))
	}

	unionCount := len(stringsSorted(append(append([]string{}, phase1Packages...), phase2Packages...)))
	thresholdCount := RunAllThresholdCount(runAllThreshold, len(allAccPackages))
	if !classified.ForceAll && len(selected) == len(allAccPackages) && unionCount > thresholdCount {
		rationale = append(rationale, fmt.Sprintf(
			"run-all threshold: union of %d packages exceeded %d (%.0f%% of %d); selecting the full suite",
			unionCount, thresholdCount, runAllThreshold, len(allAccPackages)))
	}

	plan := BuildShardPlan(selected, totalShards, minShardPackages)
	if !classified.ForceAll && classified.HasCode {
		for pkg := range phaseReasons {
			plan.Rationale = append(plan.Rationale, packageRationale(pkg, phaseReasons[pkg]))
		}
		sort.Strings(plan.Rationale)
	}
	plan.Rationale = append(rationale, plan.Rationale...)

	if err := ValidateShardPlan(plan); err != nil {
		return fmt.Errorf("validate shard plan: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(plan); err != nil {
		return fmt.Errorf("encode shard plan: %w", err)
	}
	return nil
}

func validateFlags(totalShards int, runAllThreshold float64, minShardPackages int) error {
	if totalShards <= 0 {
		return fmt.Errorf("--total-shards must be >= 1")
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

func selectionRationale(changedFiles []string) []string {
	switch {
	case len(changedFiles) == 0:
		return []string{"no resolvable diff; selecting all acceptance test packages"}
	default:
		return []string{fmt.Sprintf("diff against the resolved baseline changed %d files", len(changedFiles))}
	}
}

func packageRationale(pkg string, reasons []string) string {
	return fmt.Sprintf("%s: %s", pkg, strings.Join(reasons, "; "))
}
