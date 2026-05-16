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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDeadcodeOutput(t *testing.T) {
	t.Parallel()
	input := `
internal/pkg/foo.go:10:5: unreachable func: Foo
internal/pkg/bar.go:20:10: unreachable func: (*T).Method
some error line without match
`
	entries, err := parseDeadcodeOutput(strings.NewReader(input))
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "internal/pkg/foo.go", entries[0].file)
	assert.Equal(t, 10, entries[0].line)
	assert.Equal(t, 5, entries[0].column)
	assert.Equal(t, "Foo", entries[0].symbol)
	assert.Equal(t, "(*T).Method", entries[1].symbol)
}

func TestIntersectCandidates(t *testing.T) {
	t.Parallel()
	a := []deadcodeEntry{
		{file: "a.go", line: 1, column: 1, symbol: "Foo", packagePath: "pkg1"},
		{file: "b.go", line: 2, column: 1, symbol: "Bar", packagePath: "pkg1"},
	}
	b := []deadcodeEntry{
		{file: "a.go", line: 1, column: 1, symbol: "Foo", packagePath: "pkg1"},
		{file: "c.go", line: 3, column: 1, symbol: "Baz", packagePath: "pkg2"},
	}
	result := intersectCandidates(a, b)
	require.Len(t, result, 1)
	assert.Equal(t, "Foo", result[0].symbol)
}

func TestSelectOne_CooldownFiltering(t *testing.T) {
	t.Parallel()
	mem := &Memory{
		CooldownDays: 7,
		Attempts: []AttemptRecord{
			{Symbol: "pkg.A", AttemptedAt: time.Now().UTC().AddDate(0, 0, -1), Reason: ReasonBuildFailed},
			{Symbol: "pkg.B", AttemptedAt: time.Now().UTC().AddDate(0, 0, -10), Reason: ReasonPRCreated},
		},
	}
	candidates := []deadcodeEntry{
		{file: "a.go", line: 1, column: 1, symbol: "A", packagePath: "pkg"},
		{file: "b.go", line: 1, column: 1, symbol: "B", packagePath: "pkg"},
		{file: "c.go", line: 1, column: 1, symbol: "C", packagePath: "pkg"},
	}
	chosen := selectOne(candidates, mem, time.Now().UTC())
	require.NotNil(t, chosen)
	assert.Equal(t, "B", chosen.symbol)
}

func TestSelectOne_AllInCooldown(t *testing.T) {
	t.Parallel()
	mem := &Memory{
		CooldownDays: 7,
		Attempts: []AttemptRecord{
			{Symbol: "pkg.A", AttemptedAt: time.Now().UTC().AddDate(0, 0, -1), Reason: ReasonBuildFailed},
		},
	}
	candidates := []deadcodeEntry{
		{file: "a.go", line: 1, column: 1, symbol: "A", packagePath: "pkg"},
	}
	chosen := selectOne(candidates, mem, time.Now().UTC())
	assert.Nil(t, chosen)
}

func TestClassifyReferences_NoTests(t *testing.T) {
	t.Parallel()
	eligible, testFile := classifyReferences([]string{"internal/pkg/foo.go"})
	assert.False(t, eligible)
	assert.Empty(t, testFile)
}

func TestClassifyReferences_SingleNonAccTest(t *testing.T) {
	t.Parallel()
	eligible, testFile := classifyReferences([]string{"internal/pkg/foo_test.go"})
	assert.True(t, eligible)
	assert.Equal(t, "internal/pkg/foo_test.go", testFile)
}

func TestClassifyReferences_SingleAccTest(t *testing.T) {
	t.Parallel()
	eligible, testFile := classifyReferences([]string{"internal/pkg/acc_foo_test.go"})
	assert.False(t, eligible)
	assert.Empty(t, testFile)
}

func TestClassifyReferences_MultipleTests(t *testing.T) {
	t.Parallel()
	eligible, testFile := classifyReferences([]string{"internal/pkg/foo_test.go", "internal/pkg/bar_test.go"})
	assert.False(t, eligible)
	assert.Empty(t, testFile)
}

func TestIsInCooldown(t *testing.T) {
	t.Parallel()
	mem := &Memory{
		CooldownDays: 5,
		Attempts: []AttemptRecord{
			{Symbol: "pkg.X", AttemptedAt: time.Now().UTC().AddDate(0, 0, -3)},
		},
	}
	assert.True(t, isInCooldown(mem, "pkg.X", time.Now().UTC()))
	assert.False(t, isInCooldown(mem, "pkg.Y", time.Now().UTC()))
}

func TestRecordAndTrim(t *testing.T) {
	t.Parallel()
	mem := &Memory{Attempts: make([]AttemptRecord, 0)}
	for range maxAttempts + 10 {
		recordAttempt(mem, "pkg.A", "pkg", ReasonBuildFailed, AttemptContext{})
	}
	assert.Len(t, mem.Attempts, maxAttempts)
}

func TestSummarize(t *testing.T) {
	t.Parallel()
	mem := &Memory{
		Attempts: []AttemptRecord{
			{Symbol: "pkg.A", Package: "pkg/A", AttemptedAt: time.Now().UTC().AddDate(0, 0, -1), Reason: ReasonBuildFailed, Context: AttemptContext{}},
			{Symbol: "pkg.B", Package: "pkg/B", AttemptedAt: time.Now().UTC().AddDate(0, 0, -2), Reason: ReasonPRCreated, Context: AttemptContext{}},
			{Symbol: "pkg.C", Package: "pkg/A", AttemptedAt: time.Now().UTC().AddDate(0, 0, -3), Reason: ReasonTestsFailed, Context: AttemptContext{}},
		},
	}
	out := summarize(mem, 30)
	assert.Contains(t, out, "Total attempts: 3")
	assert.Contains(t, out, string(ReasonBuildFailed))
	assert.Contains(t, out, string(ReasonPRCreated))
	assert.Contains(t, out, "`pkg/A`: 2")
}

func TestCmdSelectMissingMemoryFlag(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := cmdSelect([]string{}, &stdout, &stderr)
	assert.Error(t, err)
}

func TestCmdRecordMissingMemoryFlag(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	err := cmdRecord([]string{"--symbol", "pkg.A", "--package", "pkg", "--reason", "pr_created"}, &stderr)
	assert.Error(t, err)
}

func TestCmdRecordAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memPath := filepath.Join(dir, "memory.json")
	var stderr bytes.Buffer
	err := cmdRecord([]string{
		"--memory", memPath,
		"--symbol", "pkg.A",
		"--package", "pkg",
		"--reason", "pr_created",
		"--context", `{"referenceFileCount":1}`,
	}, &stderr)
	require.NoError(t, err)

	mem, err := loadMemory(memPath)
	require.NoError(t, err)
	require.Len(t, mem.Attempts, 1)
	assert.Equal(t, "pkg.A", mem.Attempts[0].Symbol)
	assert.Equal(t, ReasonPRCreated, mem.Attempts[0].Reason)
	assert.Equal(t, 1, mem.Attempts[0].Context.ReferenceFileCount)
}

func TestCmdRecordInvalidReason(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memPath := filepath.Join(dir, "memory.json")
	var stderr bytes.Buffer
	err := cmdRecord([]string{
		"--memory", memPath,
		"--symbol", "pkg.A",
		"--package", "pkg",
		"--reason", "bogus",
	}, &stderr)
	assert.Error(t, err)
}

func TestCmdSummarize(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memPath := filepath.Join(dir, "memory.json")
	mem := &Memory{
		Attempts: []AttemptRecord{
			{Symbol: "pkg.A", Package: "pkg", AttemptedAt: time.Now().UTC(), Reason: ReasonBuildFailed},
		},
	}
	require.NoError(t, saveMemory(memPath, mem))

	var stdout, stderr bytes.Buffer
	err := cmdSummarize([]string{"--memory", memPath, "--days", "30"}, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Total attempts: 1")
}

func TestLoadAndSaveMemory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memPath := filepath.Join(dir, "memory.json")
	mem := &Memory{
		Version:      1,
		CooldownDays: 14,
		Attempts: []AttemptRecord{
			{Symbol: "pkg.A", Package: "pkg", AttemptedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Reason: ReasonPRCreated},
		},
	}
	require.NoError(t, saveMemory(memPath, mem))
	loaded, err := loadMemory(memPath)
	require.NoError(t, err)
	assert.Equal(t, mem.Version, loaded.Version)
	assert.Len(t, loaded.Attempts, 1)
	assert.Equal(t, ReasonPRCreated, loaded.Attempts[0].Reason)
}

func TestSaveMemoryAtomic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memPath := filepath.Join(dir, "memory.json")
	mem := &Memory{Version: 1, CooldownDays: 14}
	require.NoError(t, saveMemory(memPath, mem))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		if e.Name() != "memory.json" {
			t.Errorf("unexpected file left in dir: %s", e.Name())
		}
	}
}

func TestParseGoplsReferencesOutput(t *testing.T) {
	t.Parallel()
	input := `
internal/pkg/foo.go:42:10
internal/pkg/bar.go:100:5
internal/pkg/foo.go:55:3
:not-a-valid-line
`
	files, err := parseGoplsReferencesOutput(strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, []string{"internal/pkg/foo.go", "internal/pkg/bar.go"}, files)
}

func TestDerivePackagePath(t *testing.T) {
	t.Parallel()
	modulePath := "github.com/elastic/terraform-provider-elasticstack"
	assert.Equal(t, modulePath+"/internal/pkg", derivePackagePath("internal/pkg/foo.go", modulePath))
	assert.Equal(t, modulePath, derivePackagePath("foo.go", modulePath))
}

func TestImpactedPackages(t *testing.T) {
	t.Parallel()

	// Symbol in a sub-package, no companion test
	entry := deadcodeEntry{file: "internal/pkg/foo.go", symbol: "Foo", packagePath: "pkg"}
	pkgs := impactedPackages(entry, "")
	assert.Equal(t, []string{"./internal/pkg"}, pkgs)

	// Symbol in root package, no companion test — should yield "." not "./."
	entryRoot := deadcodeEntry{file: "foo.go", symbol: "Foo", packagePath: "root"}
	pkgsRoot := impactedPackages(entryRoot, "")
	assert.Equal(t, []string{"."}, pkgsRoot)

	// Symbol and companion test in same package
	pkgsSame := impactedPackages(entry, "internal/pkg/foo_test.go")
	assert.Equal(t, []string{"./internal/pkg"}, pkgsSame)

	// Symbol and companion test in different packages
	pkgsDiff := impactedPackages(entry, "internal/other/foo_test.go")
	assert.Equal(t, []string{"./internal/other", "./internal/pkg"}, pkgsDiff)

	// Companion test in root package
	entrySub := deadcodeEntry{file: "internal/pkg/foo.go", symbol: "Foo", packagePath: "pkg"}
	pkgsRootTest := impactedPackages(entrySub, "foo_test.go")
	assert.Equal(t, []string{".", "./internal/pkg"}, pkgsRootTest)
}
