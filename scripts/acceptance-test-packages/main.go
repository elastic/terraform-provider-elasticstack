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
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

const (
	dashboardPackage = "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard"
	lensXYPackage    = "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensxy"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "acceptance-test-packages: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("acceptance-test-packages", flag.ContinueOnError)
	flags.SetOutput(stderr)
	suite := flags.String("suite", "all", "acceptance suite: all, general, dashboard")
	stackVersion := flags.String("stack-version", "", "Stack version for general suite partitioning")
	total := flags.Int("total-shards", 1, "number of shards")
	index := flags.Int("shard-index", 0, "zero-based shard index")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *suite != "all" && *suite != "general" && *suite != "dashboard" {
		return fmt.Errorf("invalid suite %q: expected all, general, or dashboard", *suite)
	}
	if *total <= 0 || *index < 0 || *index >= *total {
		return fmt.Errorf("invalid shard assignment: total=%d index=%d (require total > 0 and 0 <= index < total)", *total, *index)
	}
	output, err := exec.Command("go", "list", "./...").Output()
	if err != nil {
		return fmt.Errorf("list packages: %w", err)
	}
	packages := selectPackages(strings.Fields(string(output)), *suite, *stackVersion, *total, *index)
	_, err = fmt.Fprintln(stdout, strings.Join(packages, " "))
	return err
}

func selectPackages(packages []string, suite, stackVersion string, total, index int) []string {
	versionParts := strings.Split(strings.TrimSuffix(stackVersion, "-SNAPSHOT"), ".")
	major, _ := strconv.Atoi(versionParts[0])
	minor := 0
	if len(versionParts) > 1 {
		minor, _ = strconv.Atoi(versionParts[1])
	}
	isolateDashboard := major > 9 || major == 9 && minor >= 4
	selected := make([]string, 0, len(packages))
	slices.Sort(packages)
	for _, pkg := range packages {
		dashboardHotspot := pkg == dashboardPackage || pkg == lensXYPackage
		if suite == "dashboard" && !dashboardHotspot || suite == "general" && isolateDashboard && dashboardHotspot {
			continue
		}
		selected = append(selected, pkg)
	}
	sharded := make([]string, 0, len(selected))
	for position, pkg := range selected {
		if position%total == index {
			sharded = append(sharded, pkg)
		}
	}
	return sharded
}
