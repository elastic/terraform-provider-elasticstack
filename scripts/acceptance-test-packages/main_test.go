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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modulePath = "github.com/elastic/terraform-provider-elasticstack"

func TestMakeDashboardSuiteContainsOnlyMeasuredHotspots(t *testing.T) {
	t.Setenv("MAKEFLAGS", "-- ACCTEST_SUITE=general ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=1 STACK_VERSION=9.3.9")
	t.Setenv("ACCTEST_SUITE", "general")
	t.Setenv("ACCTEST_TOTAL_SHARDS", "2")
	t.Setenv("ACCTEST_SHARD_INDEX", "1")
	t.Setenv("STACK_VERSION", "9.3.9")
	packages := []string{modulePath + "/internal/acctest", dashboardPackage, dashboardPackage + "/future", lensXYPackage, dashboardPackage + "_extra"}
	output, err := runMake(t, packages, "ACCTEST_SUITE=dashboard", "STACK_VERSION=9.4.0")
	require.NoError(t, err, "%s", output)
	var packageArg string
	for line := range strings.SplitSeq(string(output), "\n") {
		if strings.HasPrefix(line, "--packages=") {
			packageArg = line
		}
	}
	assert.Equal(t, "--packages="+packages[1]+" "+packages[3], packageArg)
}

func TestSupportedFullSuiteCoversEveryPackageExactlyOnce(t *testing.T) {
	for _, version := range []string{"9.4.0-SNAPSHOT", "9.10.0"} {
		t.Run(version, func(t *testing.T) {
			packages := []string{modulePath + "/z", dashboardPackage + "/future", lensXYPackage, modulePath + "/a", dashboardPackage, dashboardPackage + "_extra"}
			packageFixture(t, packages)
			var combined []string
			for _, assignment := range []struct{ suite, total, index string }{
				{"general", "2", "0"}, {"general", "2", "1"}, {"dashboard", "1", "0"},
			} {
				var output bytes.Buffer
				require.NoError(t, run([]string{"--suite=" + assignment.suite, "--stack-version=" + version, "--total-shards=" + assignment.total, "--shard-index=" + assignment.index}, &output, io.Discard))
				selected := strings.Fields(output.String())
				if assignment.suite == "general" {
					for _, pkg := range selected {
						assert.NotContains(t, []string{dashboardPackage, lensXYPackage}, pkg)
					}
				}
				combined = append(combined, selected...)
			}
			slices.Sort(packages)
			slices.Sort(combined)
			assert.Equal(t, packages, combined)
		})
	}
}

func packageFixture(t *testing.T, packages []string) {
	t.Helper()
	bin := t.TempDir()
	stub := "#!/bin/sh\nprintf '%s\\n'"
	for _, pkg := range packages {
		stub += " '" + pkg + "'"
	}
	stub += "\n"
	require.NoError(t, os.WriteFile(filepath.Join(bin, "go"), []byte(stub), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func runMake(t *testing.T, packages []string, args ...string) ([]byte, error) {
	t.Helper()
	actualGo, err := exec.LookPath("go")
	require.NoError(t, err)
	bin := t.TempDir()
	selector := filepath.Join(bin, "selector")
	build := exec.Command(actualGo, "build", "-o", selector, ".")
	buildOutput, err := build.CombinedOutput()
	require.NoError(t, err, "%s", buildOutput)
	stub := "#!/bin/sh\ncase \"$1\" in\nlist) printf '%s\\n' "
	for _, pkg := range packages {
		stub += "'" + pkg + "' "
	}
	stub += ";;\nrun) shift; shift; exec '" + selector + "' \"$@\";;\ntool) printf '%s\\n' \"$@\";;\nesac\n"
	require.NoError(t, os.WriteFile(filepath.Join(bin, "go"), []byte(stub), 0o755))
	cmd := exec.Command("make", append([]string{"testacc"}, args...)...)
	cmd.Dir = "../.."
	env := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		switch name {
		case "PATH", "MAKEFLAGS", "MFLAGS", "MAKELEVEL",
			"ACCTEST_SUITE", "ACCTEST_TOTAL_SHARDS", "ACCTEST_SHARD_INDEX", "STACK_VERSION":
			continue
		default:
			env = append(env, entry)
		}
	}
	cmd.Env = append(env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := cmd.CombinedOutput()
	return output, err
}

func TestMakeOlderAndDefaultSuitesRetainOriginalPartition(t *testing.T) {
	t.Setenv("MAKEFLAGS", "-- ACCTEST_SUITE=general ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=0 STACK_VERSION=9.5.5")
	t.Setenv("ACCTEST_SUITE", "general")
	packages := []string{modulePath + "/a", dashboardPackage, dashboardPackage + "/future", dashboardPackage + "_extra", modulePath + "/z"}
	for _, args := range [][]string{
		{"ACCTEST_SUITE=general", "STACK_VERSION=9.3.9-SNAPSHOT"},
		{"STACK_VERSION=9.4.0-SNAPSHOT"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			args = append(args, "ACCTEST_TOTAL_SHARDS=2", "ACCTEST_SHARD_INDEX=1")
			output, err := runMake(t, packages, args...)
			require.NoError(t, err, "%s", output)
			var packageArg string
			for line := range strings.SplitSeq(string(output), "\n") {
				if strings.HasPrefix(line, "--packages=") {
					packageArg = line
				}
			}
			assert.Equal(t, "--packages="+packages[1]+" "+packages[3], packageArg)
		})
	}
}

func TestMakeInvalidAssignmentFailsBeforeAcceptanceExecution(t *testing.T) {
	for _, args := range [][]string{
		{"ACCTEST_SUITE=unknown"}, {"ACCTEST_TOTAL_SHARDS=0"}, {"ACCTEST_TOTAL_SHARDS=-1"}, {"ACCTEST_SHARD_INDEX=-1"}, {"ACCTEST_TOTAL_SHARDS=2", "ACCTEST_SHARD_INDEX=2"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			output, err := runMake(t, []string{modulePath + "/a"}, args...)
			require.Error(t, err)
			assert.Contains(t, string(output), "invalid")
			assert.NotContains(t, string(output), "--packages=")
		})
	}
}
