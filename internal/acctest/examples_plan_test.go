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
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/examples"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// skippedExamplePathPrefixes lists repository-relative path prefixes excluded
// from the PlanOnly harness (REQ-005 directory skips). They are not present in the embedded
// trees ResourcesFS/DataSourcesFS; the list documents and enforces policy if
// embedding or discovery changes.
//
// - examples/cloud/: uses the Elastic Cloud (ec) provider, not elasticstack.
// - examples/provider/: partial provider-configuration snippets only.
var skippedExamplePathPrefixes = []string{
	"examples/cloud/",
	"examples/provider/",
}

// planOnlySkippedEmbedPaths lists paths under the embedded resources/ or data-sources/
// trees (same form as collectTfExamples slashPath, e.g. data-sources/foo/bar.tf) that
// cannot be planned in isolation by design. Keep this list minimal and document each entry.
var planOnlySkippedEmbedPaths = []string{
	// Requires a separate root module and terraform_remote_state; not a single-module plan.
	"data-sources/elasticstack_kibana_agentbuilder_agent/import.tf",
	// Depends on the external hashicorp/time provider; harness uses only elasticstack factories.
	"resources/elasticstack_elasticsearch_security_api_key/rotation.tf",
	// Requires a Fleet agent policy UUID that exists in Kibana/Fleet — matrix stacks yield 404 otherwise; no stack-agnostic UUID for copy-pasted examples.
	"data-sources/elasticstack_fleet_enrollment_tokens/data-source.tf",
}

func shouldSkipPlanOnlyExample(pathUnderExamples string) bool {
	return slices.Contains(planOnlySkippedEmbedPaths, pathUnderExamples)
}

func shouldSkipExamplePath(repoRelative string) bool {
	for _, prefix := range skippedExamplePathPrefixes {
		if strings.HasPrefix(repoRelative, prefix) {
			return true
		}
	}
	return false
}

// maxConcurrentExamplesPlanHarness bounds how many example PlanOnly subtests may execute
// terraform-plugin-testing workloads at once. Hundreds of unchecked t.Parallel() subtests
// each spinning a Terraform core and Configure cycle against the muxed in-process provider
// correlated with flaky refresh-plan failures where Elasticsearch resolution briefly appears
// unset ("elasticsearch client is not configured..."). A cap of 16 still reproduced flakes
// under repeated full-package runs (-count≥2); 4 aligns CI stability while keeping modest
// parallelism (t.Parallel() remains enabled).
const maxConcurrentExamplesPlanHarness = 4

var examplesPlanHarnessSem = make(chan struct{}, maxConcurrentExamplesPlanHarness)

type tfExamplePlanCase struct {
	pathUnderExamples string // e.g. resources/elasticstack_x/resource.tf — subtest id (REQ-002)
	repoExamplesPath  string // e.g. examples/resources/… — skip prefixes and resource-vs-DS semantics
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
			pathUnderExamples: slashPath,
			repoExamplesPath:  repoRel,
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

// planHarnessSubdir matches ConfigDirectory below (NamedTestCaseDirectory(planHarnessSubdirName)).
const planHarnessSubdirName = "plan"

// tfRootDeclaresResourceOrOutput returns true when the root HCL body contains a real top-level
// resource or output block (not strings/comments). This mirrors common non-empty plan causes
// when combined with ExpectNonEmptyPlan checks in terraform-plugin-testing (resource and output
// changes both contribute to a non-empty plan JSON).
func tfRootDeclaresResourceOrOutput(src []byte) bool {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(src, "example.tf")
	// Unparseable configs: return false so ExpectNonEmptyPlan is not set from a partial AST; Terraform
	// will surface invalid configuration through plan diagnostics instead.
	if diags.HasErrors() {
		return false
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return false
	}
	for _, block := range body.Blocks {
		switch block.Type {
		case "resource", "output":
			return true
		default:
		}
	}
	return false
}

// expectNonEmptyPlanForExample matches REQ-004 for data-source vs resource trees.
func expectNonEmptyPlanForExample(repoExamplesPath string, cfg []byte) bool {
	if strings.HasPrefix(repoExamplesPath, "examples/resources/") {
		return true
	}
	return tfRootDeclaresResourceOrOutput(cfg)
}

// TestAccExamples_planOnly runs each example *.tf under examples/resources and
// examples/data-sources in isolation with PlanOnly against the in-process
// provider. Subtest names are paths under examples/ (REQ-002).
// Concurrency across subtests is bounded by examplesPlanHarnessSem (see constants above).
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
		return cases[i].pathUnderExamples < cases[j].pathUnderExamples
	})

	for _, c := range cases {
		if shouldSkipPlanOnlyExample(c.pathUnderExamples) {
			continue
		}
		t.Run(c.pathUnderExamples, func(t *testing.T) {
			t.Parallel()
			examplesPlanHarnessSem <- struct{}{}
			t.Cleanup(func() { <-examplesPlanHarnessSem })

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

			expectNonEmpty := expectNonEmptyPlanForExample(c.repoExamplesPath, body)
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
