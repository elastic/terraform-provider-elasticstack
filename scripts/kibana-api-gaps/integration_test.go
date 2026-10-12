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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const repoRoot = "../.."

func TestDiscoverRealRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("real-repository discovery is skipped under -short")
	}
	ensureFilteredOAS(t)

	doc, err := loadOAS(filepath.Join(repoRoot, "generated/kbapi/oas-filtered.yaml"))
	require.NoError(t, err)
	mod, err := loadModule(repoRoot, "./...")
	require.NoError(t, err)
	resources := discoverResources()

	first, err := discoverLoaded(mod, doc, resources)
	require.NoError(t, err, "no json tag may be left unmatched")

	analysed := first.Operations
	assert.Contains(t, analysed, "elasticstack_fleet_agent_policy")
	assert.Contains(t, analysed, "elasticstack_apm_agent_configuration")

	for _, name := range entitycoreResources(t, mod, resources) {
		assert.Positive(t, analysed[name], "entitycore-based resource %s maps no operation", name)
	}

	registered := make(map[string]bool)
	for _, r := range resources {
		registered[r.Name] = true
	}
	for name := range analysed {
		assert.True(t, registered[name], "%s is not a registered resource", name)
	}
	for _, name := range first.NoOperations {
		assert.NotContains(t, analysed, name)
	}
	assert.NotEmpty(t, first.NoOperations, "Elasticsearch-only resources must be reported as not analysed")
	for _, r := range resources {
		if _, isAnalysed := analysed[r.Name]; isAnalysed {
			assert.False(t, strings.HasPrefix(r.Name, "elasticstack_elasticsearch_"), "Elasticsearch resource %s was analysed", r.Name)
			assert.NotContains(t, r.PkgPath, "/internal/elasticsearch/", "Elasticsearch package %s was analysed", r.PkgPath)
		}
	}
	assertProviderClientsPackageIsKbapiFree(t, mod)
	assert.NotEmpty(t, first.Groups)

	second, err := discoverLoaded(mod, doc, resources)
	require.NoError(t, err)
	firstJSON, err := json.Marshal(first)
	require.NoError(t, err)
	secondJSON, err := json.Marshal(second)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(firstJSON, secondJSON), "two runs must produce byte-identical output")
}

// ensureFilteredOAS generates the gitignored generated/kbapi/oas-filtered.yaml when it is absent.
func ensureFilteredOAS(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(repoRoot, "generated/kbapi/oas-filtered.yaml")); err == nil {
		return
	}
	out, err := exec.Command("make", "-C", filepath.Join(repoRoot, "generated/kbapi"), "transform").CombinedOutput()
	require.NoError(t, err, string(out))
}

func entitycoreResources(t *testing.T, mod *loadedModule, resources []Resource) []string {
	t.Helper()
	var names []string
	for _, r := range resources {
		pkg, ok := mod.byPath[r.PkgPath]
		require.True(t, ok, r.PkgPath)
		for _, obj := range pkg.TypesInfo.Uses {
			if obj.Name() == "NewKibanaResource" && obj.Pkg() != nil && strings.HasSuffix(obj.Pkg().Path(), entitycorePkgSuffix) {
				names = append(names, r.Name)
				break
			}
		}
	}
	require.NotEmpty(t, names)
	return names
}

// The reachability walk never expands into the provider's client plumbing package; this keeps
// that cut sound by asserting the package cannot hide field usage or operations.
func assertProviderClientsPackageIsKbapiFree(t *testing.T, mod *loadedModule) {
	t.Helper()
	kbapiPath, err := findPackageBySuffix(mod, kbapiPkgSuffix)
	require.NoError(t, err)
	clientsPath, err := findPackageBySuffix(mod, providerClientsPkgSuffix)
	require.NoError(t, err)
	for _, obj := range mod.byPath[clientsPath].TypesInfo.Uses {
		if obj.Pkg() != nil && obj.Pkg().Path() == kbapiPath {
			t.Errorf("%s references kbapi object %s; the reach cut at it is no longer sound", clientsPath, obj.Name())
		}
	}
}
