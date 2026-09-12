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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const providerWorkflowPath = "../../.github/workflows/provider.yml"

type providerWorkflow struct {
	Jobs map[string]providerJob `yaml:"jobs"`
}

type providerJob struct {
	Needs    yaml.Node              `yaml:"needs"`
	If       string                 `yaml:"if"`
	Outputs  map[string]string      `yaml:"outputs"`
	RunsOn   string                 `yaml:"runs-on"`
	Strategy providerStrategy       `yaml:"strategy"`
	Steps    []providerWorkflowStep `yaml:"steps"`
}

type providerStrategy struct {
	FailFast bool           `yaml:"fail-fast"`
	Matrix   map[string]any `yaml:"matrix"`
}

type providerWorkflowStep struct {
	Name string            `yaml:"name"`
	ID   string            `yaml:"id"`
	If   string            `yaml:"if"`
	Run  string            `yaml:"run"`
	Uses string            `yaml:"uses"`
	Env  map[string]string `yaml:"env"`
}

func TestProviderWorkflow_testMatrixVersionLoadedFromLoadMatrix(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	load, ok := wf.Jobs["load-matrix"]
	require.True(t, ok, "missing load-matrix job")
	assert.Equal(t, "${{ steps.load.outputs.versions }}", load.Outputs["versions"])
	assert.Equal(t, "${{ steps.load.outputs.flags }}", load.Outputs["flags"])

	var ranLoadMatrix, hasCheckout bool
	for _, step := range load.Steps {
		if strings.Contains(step.Uses, "actions/checkout@") {
			hasCheckout = true
		}
		if step.ID == "load" && step.Run == "go run ./scripts/version-matrix load-matrix" {
			ranLoadMatrix = true
		}
	}
	assert.True(t, hasCheckout, "load-matrix job must check out the repository")
	assert.True(t, ranLoadMatrix, "load-matrix job must invoke go run ./scripts/version-matrix load-matrix")
	assert.ElementsMatch(t, []string{"classify"}, providerNeeds(t, load))
	assert.Equal(t, "needs.classify.outputs.provider_changes == 'true'", load.If)

	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")
	assert.ElementsMatch(t, []string{"classify", "build", "load-matrix"}, providerNeeds(t, test))
	assert.Equal(t, "${{ fromJson(needs['load-matrix'].outputs.versions) }}", test.Strategy.Matrix["version"])
	assert.Equal(t, []any{0, 1}, test.Strategy.Matrix["shard"])
}

func TestProviderWorkflow_testStrategyHasNoIncludeOverrides(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")
	_, hasInclude := test.Strategy.Matrix["include"]
	assert.False(t, hasInclude)
}

func TestProviderWorkflow_derivedFlagsReplaceMatrixRunnerAndFleetImage(t *testing.T) {
	t.Parallel()

	raw := readWorkflowFile(t, providerWorkflowPath)
	assert.NotContains(t, raw, "needs.load-matrix")
	assert.Contains(t, raw, "needs['load-matrix']")
	assert.NotContains(t, raw, "matrix.runner")
	assert.NotContains(t, raw, "matrix.fleetImage")
	assert.NotContains(t, raw, "startsWith(matrix.version, '8.1.')")
	assert.NotContains(t, raw, "matrix.version == '8.14.3'")
	assert.NotContains(t, raw, "matrix.version == '8.15.5'")
	assert.NotContains(t, raw, "matrix.version == '8.16.6'")
	assert.NotContains(t, raw, "matrix.version == '8.17.10'")

	wf := loadProviderWorkflow(t)
	test := wf.Jobs["test"]
	assert.Equal(t, "${{ fromJson(needs['load-matrix'].outputs.flags)[matrix.version].runner }}", test.RunsOn)

	var prePull, compose, synthetics providerWorkflowStep
	for _, step := range test.Steps {
		switch {
		case step.Name == "Pre-pull fleet image":
			prePull = step
		case step.Name == "Start stack with docker compose":
			compose = step
		case step.ID == "force-install-synthetics":
			synthetics = step
		}
	}
	assert.Equal(t, "fromJson(needs['load-matrix'].outputs.flags)[matrix.version].prePullFleet", prePull.If)
	assert.Contains(t, prePull.Run, "fromJson(needs['load-matrix'].outputs.flags)[matrix.version].fleetImage")
	assert.Equal(t, "${{ fromJson(needs['load-matrix'].outputs.flags)[matrix.version].fleetImage }}", compose.Env["FLEET_IMAGE"])
	assert.Equal(t, "fromJson(needs['load-matrix'].outputs.flags)[matrix.version].forceSynthetics", synthetics.If)
}

func TestProviderWorkflow_gateInspectsLoadMatrix(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	gate, ok := wf.Jobs["gate"]
	require.True(t, ok, "missing gate job")
	assert.ElementsMatch(t, []string{"classify", "build", "golangci-lint", "lint", "test", "load-matrix"}, providerNeeds(t, gate))

	var gateStep providerWorkflowStep
	for _, step := range gate.Steps {
		if step.ID == "gate" {
			gateStep = step
			break
		}
	}
	require.NotEmpty(t, gateStep.ID)
	assert.Equal(t, "${{ needs['load-matrix'].result }}", gateStep.Env["PROVIDER_GATE_LOAD_MATRIX_RESULT"])
}

func loadProviderWorkflow(t *testing.T) providerWorkflow {
	t.Helper()
	raw := readWorkflowFile(t, providerWorkflowPath)
	var wf providerWorkflow
	require.NoError(t, yaml.Unmarshal([]byte(raw), &wf))
	return wf
}

func providerNeeds(t *testing.T, job providerJob) []string {
	t.Helper()
	switch job.Needs.Kind {
	case yaml.ScalarNode:
		return []string{job.Needs.Value}
	case yaml.SequenceNode:
		var needs []string
		require.NoError(t, job.Needs.Decode(&needs))
		return needs
	default:
		return nil
	}
}
