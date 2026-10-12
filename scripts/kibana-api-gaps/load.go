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
	"fmt"
	"strings"

	"golang.org/x/tools/go/packages"
)

// loadedModule holds the type-checked packages of one module load, shared by all analysis units.
type loadedModule struct {
	pkgs   []*packages.Package
	byPath map[string]*packages.Package
}

func loadModule(dir string, patterns ...string) (*loadedModule, error) {
	cfg := &packages.Config{
		Dir: dir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	mod := &loadedModule{pkgs: pkgs, byPath: make(map[string]*packages.Package, len(pkgs))}
	var problems []string
	for _, pkg := range pkgs {
		mod.byPath[pkg.PkgPath] = pkg
		for _, pkgErr := range pkg.Errors {
			problems = append(problems, fmt.Sprintf("%s: %s", pkg.PkgPath, pkgErr.Msg))
		}
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("load packages: %d errors, first: %s", len(problems), strings.Join(problems[:min(3, len(problems))], "; "))
	}
	return mod, nil
}
