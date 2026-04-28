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

package acctest

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/examples"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// skippedExamplePathPrefixes lists repository-relative path prefixes excluded
// from the PlanOnly harness (REQ-005). They are not present in the embedded
// trees ResourcesFS/DataSourcesFS; the list documents and enforces policy if
// embedding or discovery changes.
//
// - examples/cloud/: uses the Elastic Cloud (ec) provider, not elasticstack.
// - examples/provider/: partial provider-configuration snippets only.
var skippedExamplePathPrefixes = []string{
	"examples/cloud/",
	"examples/provider/",
}

func shouldSkipExamplePath(repoRelative string) bool {
	for _, prefix := range skippedExamplePathPrefixes {
		if strings.HasPrefix(repoRelative, prefix) {
			return true
		}
	}
	return false
}

type tfExamplePlanCase struct {
	repoRelative      string // e.g. examples/resources/elasticstack_x/resource.tf
	fsys              fs.FS
	embedRelativePath string // path within ResourcesFS/DataSourcesFS
}

func collectTfExamples(fsys fs.FS) ([]tfExamplePlanCase, error) {
	var out []tfExamplePlanCase
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tf") {
			return nil
		}
		// Embed paths begin with resources/… or data-sources/… (directory name preserved).
		slashPath := filepath.ToSlash(path)
		repoRel := "examples/" + slashPath
		if shouldSkipExamplePath(repoRel) {
			return nil
		}
		out = append(out, tfExamplePlanCase{
			repoRelative:      repoRel,
			fsys:              fsys,
			embedRelativePath: path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// planHarnessSubdir matches ConfigDirectory below (NamedTestCaseDirectory(planHarnessDirName)).
const planHarnessSubdirName = "plan"

// expectNonEmptyPlanForExample matches REQ-004 resource vs data-directory expectations against
// terraform-plugin-testing semantics for PlanOnly steps: both the non-refresh and refresh-plan
// checks key off ExpectNonEmptyPlan — use false only when plans are expected empty (typically
// no managed resources declared in that file).
func expectNonEmptyPlanForExample(repoRelative string, cfg []byte) bool {
	if strings.HasPrefix(repoRelative, "examples/resources/") {
		return true
	}
	// Data-source docs tree: tolerate both read-only-only (empty plan) and configs that declare
	// prerequisite managed resources alongside data sources / data blocks.
	return bytes.Contains(cfg, []byte(`resource "`))
}

// TestAccExamples_planOnly runs each example *.tf under examples/resources and
// examples/data-sources in isolation with PlanOnly against the in-process
// provider. Subtest names are the repo-relative paths (REQ-002).
func TestAccExamples_planOnly(t *testing.T) {
	var cases []tfExamplePlanCase
	res, err := collectTfExamples(examples.ResourcesFS)
	if err != nil {
		t.Fatalf("walk resources examples: %v", err)
	}
	cases = append(cases, res...)

	ds, err := collectTfExamples(examples.DataSourcesFS)
	if err != nil {
		t.Fatalf("walk data-sources examples: %v", err)
	}
	cases = append(cases, ds...)

	sort.Slice(cases, func(i, j int) bool {
		return cases[i].repoRelative < cases[j].repoRelative
	})

	for _, c := range cases {
		t.Run(c.repoRelative, func(t *testing.T) {
			t.Parallel()

			body, err := fs.ReadFile(c.fsys, c.embedRelativePath)
			if err != nil {
				t.Fatalf("read embedded %s: %v", c.embedRelativePath, err)
			}

			testDataBranch := filepath.Join("testdata", t.Name())
			planDir := filepath.Join(testDataBranch, planHarnessSubdirName)
			t.Cleanup(func() {
				if errRemove := os.RemoveAll(testDataBranch); errRemove != nil {
					t.Logf("remove generated testdata %s: %v", testDataBranch, errRemove)
				}
			})
			if err := os.MkdirAll(planDir, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", planDir, err)
			}
			tfName := filepath.Base(c.embedRelativePath)
			tfPath := filepath.Join(planDir, tfName)
			if err := os.WriteFile(tfPath, body, 0o644); err != nil {
				t.Fatalf("write %s: %v", tfPath, err)
			}

			expectNonEmpty := expectNonEmptyPlanForExample(c.repoRelative, body)
			resource.Test(t, resource.TestCase{
				PreCheck: func() { PreCheck(t) },
				Steps: []resource.TestStep{
					{
						ProtoV6ProviderFactories: Providers,
						ConfigDirectory:          NamedTestCaseDirectory(planHarnessSubdirName),
						PlanOnly:                 true,
						ExpectNonEmptyPlan:       expectNonEmpty,
					},
				},
			})
		})
	}
}
