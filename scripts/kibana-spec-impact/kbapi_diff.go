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
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const kbapiGenPath = "generated/kbapi/kibana.gen.go"

// kbapiSurface returns exported type names and ClientWithResponses method names from kibana.gen.go source.
func kbapiSurface(src string) (types, methods []string, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "kibana.gen.go", src, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parse kbapi: %w", err)
	}

	typeSet := make(map[string]struct{})
	methodSet := make(map[string]struct{})

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				if ts.Name.IsExported() {
					typeSet[ts.Name.Name] = struct{}{}
				}
			}
		case *ast.FuncDecl:
			if d.Recv == nil || len(d.Recv.List) != 1 {
				continue
			}
			recv := d.Recv.List[0].Type
			star, ok := recv.(*ast.StarExpr)
			if !ok {
				continue
			}
			id, ok := star.X.(*ast.Ident)
			if !ok || id.Name != "ClientWithResponses" {
				continue
			}
			if d.Name != nil && d.Name.IsExported() {
				methodSet[d.Name.Name] = struct{}{}
			}
		}
	}

	for k := range typeSet {
		types = append(types, k)
	}
	for k := range methodSet {
		methods = append(methods, k)
	}
	sort.Strings(types)
	sort.Strings(methods)
	return types, methods, nil
}

// diffKbapiSurfaces returns symbol names that were added or removed between two versions of kibana.gen.go.
// normalizeKbapiSourceForDiff maps missing or empty generated file content to a minimal parseable
// package so surface extraction does not fail when kibana.gen.go is absent at one revision.
func normalizeKbapiSourceForDiff(src string) string {
	if strings.TrimSpace(src) == "" {
		return "package kbapi\n"
	}
	return src
}

func diffKbapiSurfaces(oldSrc, newSrc string) ([]string, error) {
	oldSrc = normalizeKbapiSourceForDiff(oldSrc)
	newSrc = normalizeKbapiSourceForDiff(newSrc)
	oldTypes, oldMethods, err := kbapiSurface(oldSrc)
	if err != nil {
		return nil, fmt.Errorf("old kbapi surface: %w", err)
	}
	newTypes, newMethods, err := kbapiSurface(newSrc)
	if err != nil {
		return nil, fmt.Errorf("new kbapi surface: %w", err)
	}

	oldSet := make(map[string]struct{})
	for _, s := range append(append([]string{}, oldTypes...), oldMethods...) {
		oldSet[s] = struct{}{}
	}
	newSet := make(map[string]struct{})
	for _, s := range append(append([]string{}, newTypes...), newMethods...) {
		newSet[s] = struct{}{}
	}

	var changed []string
	for s := range newSet {
		if _, ok := oldSet[s]; !ok {
			changed = append(changed, s)
		}
	}
	for s := range oldSet {
		if _, ok := newSet[s]; !ok {
			changed = append(changed, s)
		}
	}
	sort.Strings(changed)
	return changed, nil
}

// gitShowPathOrMissing returns file content at rev:path, or missing=true when the path does not
// exist in that revision (add/remove/rename of generated/kbapi/kibana.gen.go).
func gitShowPathOrMissing(repoRoot, rev string) (content string, missing bool, err error) {
	repoRoot = filepath.Clean(repoRoot)
	cmd := exec.Command("git", "-C", repoRoot, "show", rev+":"+kbapiGenPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err == nil {
		return stdout.String(), false, nil
	}
	msg := stderr.String()
	if strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "exists on disk, but not in") ||
		strings.Contains(msg, "did not match any file") {
		return "", true, nil
	}
	return "", false, fmt.Errorf("git show %s:%s: %w\n%s", rev, kbapiGenPath, err, msg)
}

func diffKbapiAtRefs(repoRoot, baselineRev, targetRev string) ([]string, error) {
	repoRoot = filepath.Clean(repoRoot)
	oldSrc, oldMiss, err := gitShowPathOrMissing(repoRoot, baselineRev)
	if err != nil {
		return nil, err
	}
	newSrc, newMiss, err := gitShowPathOrMissing(repoRoot, targetRev)
	if err != nil {
		return nil, err
	}
	if oldMiss {
		oldSrc = ""
	}
	if newMiss {
		newSrc = ""
	}
	return diffKbapiSurfaces(oldSrc, newSrc)
}

func gitRevParse(repoRoot, rev string) (string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", rev)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse %s: %w", rev, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
