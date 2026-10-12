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
	doc, err := loadOAS(filepath.Join(root, "generated/kbapi/oas-filtered.yaml"))
	if err != nil {
		return nil, err
	}
	mod, err := loadModule(root, "./...")
	if err != nil {
		return nil, err
	}
	kbapiPath, err := findKbapiPackage(mod)
	if err != nil {
		return nil, err
	}
	var resources []Resource
	for _, r := range discoverResources() {
		if len(only) == 0 || slices.Contains(only, r.Name) {
			resources = append(resources, r)
		}
	}
	return discover(discoverConfig{module: mod, kbapiPath: kbapiPath, oas: doc, resources: resources})
}

func findKbapiPackage(mod *loadedModule) (string, error) {
	for path := range mod.byPath {
		if len(path) > len(kbapiPkgSuffix) && path[len(path)-len(kbapiPkgSuffix):] == kbapiPkgSuffix {
			return path, nil
		}
	}
	return "", errors.New("generated/kbapi package not found in the module")
}
