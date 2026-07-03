// Package main implements a targeted acceptance test package selector.
package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// ImportGraph exposes the forward import relationships for internal packages.
type ImportGraph struct {
	Forward map[string][]string
	Reverse map[string][]string
}

// BuildImportGraph runs go list to obtain the non-test import graph for all
// packages under ./internal/... It returns both forward and reverse forms.
func BuildImportGraph() (*ImportGraph, error) {
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./internal/...")
	out, err := cmd.Output()
	if err != nil {
		if xerr, ok := err.(*exec.ExitError); ok && len(xerr.Stderr) > 0 {
			return nil, fmt.Errorf("go list failed: %w\n%s", err, xerr.Stderr)
		}
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	g := &ImportGraph{
		Forward: make(map[string][]string),
		Reverse: make(map[string][]string),
	}

	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pkg := fields[0]
		imports := fields[1:]
		g.Forward[pkg] = imports
		for _, imp := range imports {
			g.Reverse[imp] = append(g.Reverse[imp], pkg)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	for pkg := range g.Forward {
		sort.Strings(g.Forward[pkg])
		g.Forward[pkg] = uniqStrings(g.Forward[pkg])
	}
	for pkg := range g.Reverse {
		sort.Strings(g.Reverse[pkg])
		g.Reverse[pkg] = uniqStrings(g.Reverse[pkg])
	}

	return g, nil
}

// BuildReverseDepGraph builds a reverse dependency graph from a forward graph.
// Each key maps to the set of packages that directly import it.
func BuildReverseDepGraph(forward map[string][]string) map[string][]string {
	reverse := make(map[string][]string)
	for pkg, imports := range forward {
		for _, imp := range imports {
			reverse[imp] = append(reverse[imp], pkg)
		}
	}
	for pkg := range reverse {
		sort.Strings(reverse[pkg])
		reverse[pkg] = uniqStrings(reverse[pkg])
	}
	return reverse
}

// WalkReverseDeps returns all unique import paths that transitively import any
// package in start, including the start packages themselves. The supplied
// reverse map must map a package import path to the import paths that directly
// import it.
func WalkReverseDeps(reverse map[string][]string, start []string) []string {
	seen := make(map[string]struct{})
	for _, pkg := range start {
		seen[pkg] = struct{}{}
	}

	queue := make([]string, len(start))
	copy(queue, start)

	for len(queue) > 0 {
		pkg := queue[0]
		queue = queue[1:]
		for _, importer := range reverse[pkg] {
			if _, ok := seen[importer]; ok {
				continue
			}
			seen[importer] = struct{}{}
			queue = append(queue, importer)
		}
	}

	result := make([]string, 0, len(seen))
	for pkg := range seen {
		result = append(result, pkg)
	}
	sort.Strings(result)
	return result
}

func stringsSorted(s []string) []string {
	cpy := append([]string(nil), s...)
	sort.Strings(cpy)
	return uniqStrings(cpy)
}

func uniqStrings(sorted []string) []string {
	if len(sorted) == 0 {
		return sorted
	}
	uniq := make([]string, 1, len(sorted))
	uniq[0] = sorted[0]
	for _, s := range sorted[1:] {
		if s == uniq[len(uniq)-1] {
			continue
		}
		uniq = append(uniq, s)
	}
	return uniq
}
