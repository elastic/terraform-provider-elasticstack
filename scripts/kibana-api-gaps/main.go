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

// Command kibana-api-gaps finds Kibana API request fields that Terraform resources never use.
//
// Usage:
//
//	go run ./scripts/kibana-api-gaps discover [--repo .] [--resource elasticstack_kibana_slo]
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const kbapiPkgSuffix = "/generated/kbapi"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "discover" {
		fmt.Fprintln(stderr, "Usage: kibana-api-gaps discover [--repo dir] [--resource name]...")
		return errors.New("unknown or missing command")
	}
	return cmdDiscover(args[1:], stdout, stderr)
}

type stringList []string

func (s *stringList) String() string     { return fmt.Sprint([]string(*s)) }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

func cmdDiscover(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", ".", "repository root")
	var only stringList
	fs.Var(&only, "resource", "restrict discovery to this resource type name (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := discoverRepository(*repo, only)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// discoverRepository runs discovery over the provider and generated/kbapi/oas-filtered.yaml
// of the repository at root, optionally restricted to the named resources.
func discoverRepository(root string, only []string) (*Discovery, error) {
	resources, err := selectResources(discoverResources(), only)
	if err != nil {
		return nil, err
	}
	doc, err := loadOAS(filepath.Join(root, "generated/kbapi/oas-filtered.yaml"))
	if err != nil {
		return nil, err
	}
	mod, err := loadModule(root, "./...")
	if err != nil {
		return nil, err
	}
	return discoverLoaded(mod, doc, resources)
}

func discoverLoaded(mod *loadedModule, doc *oasDoc, resources []Resource) (*Discovery, error) {
	kbapiPath, err := findPackageBySuffix(mod, kbapiPkgSuffix)
	if err != nil {
		return nil, err
	}
	return discover(discoverConfig{module: mod, kbapiPath: kbapiPath, oas: doc, resources: resources})
}

// selectResources restricts resources to the named ones; an unknown name is an error so a
// typo cannot silently produce an empty report.
func selectResources(all []Resource, only []string) ([]Resource, error) {
	if len(only) == 0 {
		return all, nil
	}
	var selected []Resource
	for _, name := range only {
		i := slices.IndexFunc(all, func(r Resource) bool { return r.Name == name })
		if i < 0 {
			return nil, fmt.Errorf("unknown resource %q", name)
		}
		selected = append(selected, all[i])
	}
	return selected, nil
}

func findPackageBySuffix(mod *loadedModule, suffix string) (string, error) {
	var matches []string
	for path := range mod.byPath {
		if strings.HasSuffix(path, suffix) {
			matches = append(matches, path)
		}
	}
	if len(matches) != 1 {
		slices.Sort(matches)
		return "", fmt.Errorf("expected exactly one package ending in %s, found %v", suffix, matches)
	}
	return matches[0], nil
}
