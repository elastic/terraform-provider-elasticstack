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
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

const (
	versionMatrixWorkflowPath = "../../.github/workflows/version-matrix-generation.yml"
	changelogWorkflowPath     = "../../.github/workflows/changelog-generation.yml"
)

type workflowYAML struct {
	On          workflowOn             `yaml:"on"`
	Permissions workflowPermissions    `yaml:"permissions"`
	Jobs        map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
	Name string            `yaml:"name"`
	ID   string            `yaml:"id"`
	If   string            `yaml:"if"`
	Run  string            `yaml:"run"`
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
}

type workflowOn struct {
	Schedule         []workflowSchedule `yaml:"schedule"`
	WorkflowDispatch *struct{}          `yaml:"workflow_dispatch"`
}

type workflowSchedule struct {
	Cron string `yaml:"cron"`
}

type workflowPermissions struct {
	Contents     string `yaml:"contents"`
	PullRequests string `yaml:"pull-requests"`
}

func TestVersionMatrixWorkflow_pushRefMatchesStandingHeadBranch(t *testing.T) {
	t.Parallel()

	push := workflowStepByID(t, "push")
	assert.Contains(t, push.Run, "git push origin HEAD:"+standingPRHeadBranch+" --force")
}

func TestVersionMatrixWorkflow_gitAddMatchesDefaultArtifactPath(t *testing.T) {
	t.Parallel()

	push := workflowStepByID(t, "push")
	assert.Contains(t, push.Run, "git add "+defaultArtifactPath)
}

func TestVersionMatrixWorkflow_stepGatesAndCheckoutRef(t *testing.T) {
	t.Parallel()

	wf := loadWorkflow(t, versionMatrixWorkflowPath)
	job, ok := wf.Jobs["generate-version-matrix"]
	require.True(t, ok, "missing generate-version-matrix job")

	var checkout workflowStep
	for _, step := range job.Steps {
		if strings.Contains(step.Uses, "actions/checkout@") {
			checkout = step
			break
		}
	}
	require.NotEmpty(t, checkout.Uses)
	assert.Equal(t, "${{ github.event.repository.default_branch }}", checkout.With["ref"])

	compute := workflowStepByID(t, "compute")
	assert.Empty(t, compute.If)
	assert.Equal(t, "go run ./scripts/version-matrix compute", strings.TrimSpace(compute.Run))

	push := workflowStepByID(t, "push")
	assert.Equal(t, "steps.compute.outputs.changed == 'true'", push.If)

	lookup := workflowStepByID(t, "lookup_pr")
	assert.Equal(t, "steps.push.outputs.pushed == 'true'", lookup.If)
	assert.Contains(t, lookup.Run, "go run ./scripts/version-matrix manage-pr")
}

func TestVersionMatrixWorkflow_invokesGoComputeEngine(t *testing.T) {
	t.Parallel()

	raw := readWorkflowFile(t, versionMatrixWorkflowPath)
	assert.Contains(t, raw, "actions/checkout@")
	assert.Contains(t, raw, "actions/setup-go@")
	assert.Contains(t, raw, "go run ./scripts/version-matrix compute")
}

func TestVersionMatrixWorkflow_managePRRunsOnlyWhenPushed(t *testing.T) {
	t.Parallel()

	lookup := workflowStepByID(t, "lookup_pr")
	assert.Contains(t, lookup.Run, "go run ./scripts/version-matrix manage-pr")
	assert.Equal(t, "steps.push.outputs.pushed == 'true'", lookup.If)
}

func TestVersionMatrixWorkflow_changedPushesEmptyCITriggerCommit(t *testing.T) {
	t.Parallel()

	raw := readWorkflowFile(t, versionMatrixWorkflowPath)
	assert.Contains(t, raw, "GH_AW_CI_TRIGGER_TOKEN")
	assert.Contains(t, raw, `git commit --allow-empty -m "chore: trigger CI"`)
}

func TestVersionMatrixWorkflow_changedCommitsBotAuthoredBranch(t *testing.T) {
	t.Parallel()

	raw := readWorkflowFile(t, versionMatrixWorkflowPath)
	assert.Contains(t, raw, `git config user.name "github-actions[bot]"`)
	assert.Contains(t, raw, `git config user.email "github-actions[bot]@users.noreply.github.com"`)
	assert.Contains(t, raw, "git push origin HEAD:"+standingPRHeadBranch+" --force")
}

func TestVersionMatrixWorkflow_pushRunsOnlyWhenChanged(t *testing.T) {
	t.Parallel()

	push := workflowStepByID(t, "push")
	assert.Equal(t, "steps.compute.outputs.changed == 'true'", push.If)
}

func TestVersionMatrixWorkflow_hasDailyScheduleAndDispatch(t *testing.T) {
	t.Parallel()

	got := loadWorkflow(t, versionMatrixWorkflowPath)
	changelog := loadWorkflow(t, changelogWorkflowPath)

	require.NotNil(t, got.On.WorkflowDispatch)
	require.Len(t, got.On.Schedule, 1)
	assert.Regexp(t, `^\S+ \S+ \* \* \*$`, got.On.Schedule[0].Cron)
	require.NotEmpty(t, changelog.On.Schedule)
	assert.NotEqual(t, changelog.On.Schedule[0].Cron, got.On.Schedule[0].Cron)
	assert.Equal(t, "write", got.Permissions.Contents)
	assert.Equal(t, "write", got.Permissions.PullRequests)
}

func workflowStepByID(t *testing.T, id string) workflowStep {
	t.Helper()
	wf := loadWorkflow(t, versionMatrixWorkflowPath)
	job, ok := wf.Jobs["generate-version-matrix"]
	require.True(t, ok, "missing generate-version-matrix job")
	for _, step := range job.Steps {
		if step.ID == id {
			return step
		}
	}
	t.Fatalf("missing workflow step id %q", id)
	return workflowStep{}
}

func loadWorkflow(t *testing.T, path string) workflowYAML {
	t.Helper()
	raw := readWorkflowFile(t, path)
	var wf workflowYAML
	require.NoError(t, yaml.Unmarshal([]byte(raw), &wf))
	return wf
}

func readWorkflowFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(raw)
}
