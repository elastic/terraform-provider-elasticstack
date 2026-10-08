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
	Name            string            `yaml:"name"`
	ID              string            `yaml:"id"`
	If              string            `yaml:"if"`
	Run             string            `yaml:"run"`
	Uses            string            `yaml:"uses"`
	Env             map[string]string `yaml:"env"`
	TimeoutMinutes  int               `yaml:"timeout-minutes"`
	ContinueOnError string            `yaml:"continue-on-error"`
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
	assert.Equal(t, "${{ fromJson(needs['load-matrix'].outputs.shards) }}", test.Strategy.Matrix["shard"])
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
	assert.Equal(t, "fromJson(needs['load-matrix'].outputs.flags)[matrix.version].prePullFleet && needs['load-matrix'].outputs.has_packages == 'true'", prePull.If)
	assert.Contains(t, prePull.Run, "fromJson(needs['load-matrix'].outputs.flags)[matrix.version].fleetImage")
	assert.Equal(t, "${{ fromJson(needs['load-matrix'].outputs.flags)[matrix.version].fleetImage }}", compose.Env["FLEET_IMAGE"])
	assert.Equal(t, "needs['load-matrix'].outputs.has_packages == 'true' && fromJson(needs['load-matrix'].outputs.flags)[matrix.version].forceSynthetics", synthetics.If)
}

func TestProviderWorkflow_gateInspectsLoadMatrix(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	gate, ok := wf.Jobs["gate"]
	require.True(t, ok, "missing gate job")
	assert.ElementsMatch(t, []string{"classify", "build", "golangci-lint", "lint", "test", "load-matrix", "unit-test"}, providerNeeds(t, gate))

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

func TestProviderWorkflow_loadMatrixPreparesTargetedPlan(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	load, ok := wf.Jobs["load-matrix"]
	require.True(t, ok, "missing load-matrix job")

	assert.Equal(t, "${{ steps.load.outputs.versions }}", load.Outputs["versions"])
	assert.Equal(t, "${{ steps.load.outputs.flags }}", load.Outputs["flags"])
	assert.Equal(t, "${{ steps.targeted.outputs.has_packages }}", load.Outputs["has_packages"])
	assert.Equal(t, "${{ steps.targeted.outputs.shards }}", load.Outputs["shards"])
	assert.Equal(t, "${{ steps.targeted.outputs.packages_0 }}", load.Outputs["packages_0"])
	assert.Equal(t, "${{ steps.targeted.outputs.packages_1 }}", load.Outputs["packages_1"])

	var prepare providerWorkflowStep
	for _, step := range load.Steps {
		if step.ID == "targeted" {
			prepare = step
			break
		}
	}
	require.NotEmpty(t, prepare.ID, "load-matrix job must prepare the targeted plan once")
	assert.Equal(t, "./.github/scripts/targeted-testacc-compute.sh", prepare.Run)
	assert.Equal(t, "${{ github.event_name }}", prepare.Env["EVENT_NAME"])
	assert.Equal(t, "${{ github.event.pull_request.base.sha }}", prepare.Env["PR_BASE_SHA"])

	// The test job must not re-run selection; it consumes the prepared plan.
	raw := readWorkflowFile(t, providerWorkflowPath)
	assert.NotContains(t, raw, "--shard-index")
	testJob, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")
	for _, step := range testJob.Steps {
		assert.NotContains(t, step.If, "steps.targeted.outputs")
		assert.NotContains(t, step.Run, "steps.targeted.outputs")
	}
}

func TestProviderWorkflow_testJobRunsPreparedPackages(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")

	for _, step := range test.Steps {
		assert.NotEqual(t, "targeted", step.ID,
			"test job must not compute targeted packages; preparation owns selection")
	}

	var tfAcceptance providerWorkflowStep
	for _, step := range test.Steps {
		if step.ID == "tf-acceptance" {
			tfAcceptance = step
			break
		}
	}
	require.NotEmpty(t, tfAcceptance.ID, "missing tf-acceptance step")
	assert.Contains(t, tfAcceptance.Run,
		`make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"`)
	assert.NotContains(t, tfAcceptance.Run, "go run ./scripts/targeted-testacc")
	assert.Contains(t, tfAcceptance.Run, "make testacc ACCTEST_TOTAL_SHARDS=2")
	assert.Contains(t, tfAcceptance.Run, "ACCTEST_SHARD_INDEX=${{ matrix.shard }}")
	assert.Equal(
		t,
		"${{ needs['load-matrix'].outputs[format('packages_{0}', matrix.shard)] }}",
		tfAcceptance.Env["TARGETED_PKGS"],
	)
}

func TestProviderWorkflow_nonPullRequestRunsFixedFullSuiteShards(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")

	tfAcceptance := providerWorkflowStepByID(t, test, "tf-acceptance")
	assert.Contains(t, tfAcceptance.Run,
		`"${{ github.event_name }}" == "pull_request"`,
		"the full-suite/targeted routing must branch on the event name")

	prBranch, fullSuiteBranch := splitConditionalRunBranches(t, tfAcceptance.Run)

	assert.Contains(t, prBranch, `make targeted-testacc TARGETED_PKGS="$TARGETED_PKGS"`)
	assert.NotContains(t, prBranch, "ACCTEST_TOTAL_SHARDS",
		"the prepared plan already assigned packages; no further sharding")
	assert.NotContains(t, prBranch, "go run ./scripts/targeted-testacc",
		"selection is prepared once before matrix fan-out")

	// Push, workflow-dispatch, and merge-group events keep the pre-change
	// fixed two-shard full-suite invocation, untouched by targeted planning.
	assert.Contains(t, fullSuiteBranch, "make testacc ACCTEST_TOTAL_SHARDS=2")
	assert.Contains(t, fullSuiteBranch, "ACCTEST_SHARD_INDEX=${{ matrix.shard }}")
	assert.NotContains(t, fullSuiteBranch, "TARGETED_PKGS",
		"non-pull-request events must not use targeted planning")
	assert.NotContains(t, fullSuiteBranch, "targeted-testacc",
		"non-pull-request events must not use targeted planning")
}

func TestProviderWorkflow_testJobSkipsCostlyStepsWithoutPackages(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")

	gated := "needs['load-matrix'].outputs.has_packages == 'true'"
	// Every costly step — disk cleanup, checkout/toolchain, dependencies,
	// stack lifecycle, and test execution — must be gated so a zero-package
	// pull request yields a successful no-op job. The snapshot-warning step
	// is the one exception: it keys off the test step's outcome instead.
	for _, step := range test.Steps {
		if step.Name == "Warn PR for snapshot acceptance failure" {
			continue
		}
		assert.Contains(t, step.If, gated,
			"step %q must be gated on the prepared has_packages result", step.Name)
	}

	// Checkout and toolchain setup are skipped for a no-op job too.
	for _, step := range test.Steps {
		if strings.Contains(step.Uses, "actions/checkout@") {
			assert.Contains(t, step.If, gated,
				"test-job checkout step %q must gate on the prepared plan", step.Name)
		}
	}
}

func TestProviderWorkflow_unitTestIndependentOfPreparedPlan(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	unit, ok := wf.Jobs["unit-test"]
	require.True(t, ok, "missing unit-test job")
	assert.ElementsMatch(t, []string{"classify"}, providerNeeds(t, unit))
	assert.Equal(t, "needs.classify.outputs.provider_changes == 'true'", unit.If)
	for _, step := range unit.Steps {
		assert.NotContains(t, step.Run, "has_packages")
	}
}

func TestProviderWorkflow_fullSuiteFailureHandlingPreserved(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")

	// The matrix is non-fail-fast and fixed two-shard full-suite execution
	// survives unchanged for push, workflow-dispatch, and merge-group events.
	assert.False(t, test.Strategy.FailFast,
		"matrix acceptance jobs must not fail fast")
	assert.Equal(t, "${{ fromJson(needs['load-matrix'].outputs.shards) }}",
		test.Strategy.Matrix["shard"],
		"the shard axis comes from the prepared plan")

	tfAcceptance := providerWorkflowStepByID(t, test, "tf-acceptance")
	assert.Equal(t, 45, tfAcceptance.TimeoutMinutes,
		"the acceptance step timeout must not regress")
	assert.Equal(t, "${{ endsWith(matrix.version, '-SNAPSHOT') }}",
		tfAcceptance.ContinueOnError,
		"snapshot versions are allowed to fail while non-snapshot versions block")

	// Compose startup must time out before the job-level timeout so a hung
	// pull fails the job before the runner reaps it.
	var compose providerWorkflowStep
	for _, step := range test.Steps {
		if step.Name == "Start stack with docker compose" {
			compose = step
			break
		}
	}
	require.NotEmpty(t, compose.Name)
	assert.Equal(t, 10, compose.TimeoutMinutes)
	assert.Less(t, compose.TimeoutMinutes, tfAcceptance.TimeoutMinutes)
	assert.Contains(t, compose.Run, "make docker-fleet")
}

func TestProviderWorkflow_diagnosticsTeardownAndPrePullPreserved(t *testing.T) {
	t.Parallel()

	wf := loadProviderWorkflow(t)
	test, ok := wf.Jobs["test"]
	require.True(t, ok, "missing test job")

	var prePull, diagnostics, teardown providerWorkflowStep
	for _, step := range test.Steps {
		switch step.Name {
		case "Pre-pull fleet image":
			prePull = step
		case "Docker compose logs":
			diagnostics = step
		case "Tear down docker compose stack":
			teardown = step
		}
	}
	require.NotEmpty(t, prePull.Name)
	require.NotEmpty(t, diagnostics.Name)
	require.NotEmpty(t, teardown.Name)

	// Pre-pull keeps its per-attempt timeout and 3-attempt retry loop.
	assert.Equal(t, 5, prePull.TimeoutMinutes)
	assert.Contains(t, prePull.Run, "timeout 90 docker pull")
	assert.Contains(t, prePull.Run, "for i in {1..3}")

	assert.Contains(t, diagnostics.If, "failure() || steps.tf-acceptance.outcome == 'failure'")
	assert.Contains(t, diagnostics.If, "needs['load-matrix'].outputs.has_packages == 'true'")

	assert.Contains(t, teardown.If, "always()")
	assert.Contains(t, teardown.If, "needs['load-matrix'].outputs.has_packages == 'true'")
	assert.Equal(t, "make docker-clean", teardown.Run)
}

func loadProviderWorkflow(t *testing.T) providerWorkflow {
	t.Helper()
	raw := readWorkflowFile(t, providerWorkflowPath)
	var wf providerWorkflow
	require.NoError(t, yaml.Unmarshal([]byte(raw), &wf))
	return wf
}

func providerWorkflowStepByID(t *testing.T, job providerJob, id string) providerWorkflowStep {
	t.Helper()
	for _, step := range job.Steps {
		if step.ID == id {
			return step
		}
	}
	t.Fatalf("missing step with id %q in job steps", id)
	return providerWorkflowStep{}
}

// splitConditionalRunBranches splits a bash run script's if/else body into the
// pull-request branch and the full-suite branch, normalizing line wrapping.
func splitConditionalRunBranches(t *testing.T, run string) (prBranch, fullSuiteBranch string) {
	t.Helper()
	lines := strings.Split(run, "\n")
	var branchStart int
	for i, line := range lines {
		if strings.Contains(line, "if [[") {
			branchStart = i + 1
			break
		}
	}
	require.NotZero(t, branchStart, "run script must branch on the pull-request event")

	var prLines, fullSuiteLines []string
	inElse := false
	for _, line := range lines[branchStart:] {
		switch {
		case strings.Contains(line, "else"):
			inElse = true
		case strings.Contains(line, "fi"):
			return joinRunLines(prLines), joinRunLines(fullSuiteLines)
		case inElse:
			fullSuiteLines = append(fullSuiteLines, line)
		default:
			prLines = append(prLines, line)
		}
	}
	t.Fatal("run script if/else is not closed with fi")
	return "", ""
}

func joinRunLines(lines []string) string {
	var joined strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		joined.WriteString(trimmed)
		joined.WriteByte(' ')
	}
	return strings.TrimRight(joined.String(), " ")
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
