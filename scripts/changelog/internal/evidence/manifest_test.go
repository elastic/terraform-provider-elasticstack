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

package evidence_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/evidence"
)

func TestBuildEvidenceArtifactPlan_manifest_nil(t *testing.T) {
	t.Parallel()
	_, err := evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{Manifest: nil})
	if err == nil || !strings.Contains(err.Error(), "manifest must be a non-null object") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBuildEvidenceArtifactPlan_manifest_not_object(t *testing.T) {
	t.Parallel()
	_, err := evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{Manifest: []any{}})
	if err == nil || !strings.Contains(err.Error(), "manifest must be a non-null object") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBuildEvidenceArtifactPlan_manifest_scalar_rejected(t *testing.T) {
	t.Parallel()
	for _, m := range []any{"oops", float64(1), false} {
		_, err := evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{Manifest: m})
		if err == nil || !strings.Contains(err.Error(), "manifest must be a non-null object") {
			t.Fatalf("manifest %+v unexpected err: %v", m, err)
		}
	}
}

func TestBuildEvidenceManifest_nilEvidence_sliceMarshalsPullRequests(t *testing.T) {
	t.Parallel()
	genAt := time.Date(2026, 4, 17, 8, 0, 0, 0, time.UTC)
	got := evidence.BuildEvidenceManifest("unreleased", "", "", "HEAD", nil, genAt)
	raw, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"pull_requests": null`) {
		t.Fatalf("wanted empty JSON array not null:\n%s", raw)
	}
	if !strings.Contains(string(raw), `"pull_requests": []`) {
		t.Fatalf(`missing \"pull_requests\": [] in JSON:\n%s`, raw)
	}
}

func TestBuildEvidenceArtifactPlan_requires_artifact_meta(t *testing.T) {
	t.Parallel()

	blankName := ""
	_, err := evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{
		Manifest:     map[string]any{},
		ArtifactName: &blankName,
	})
	if err == nil || !strings.Contains(err.Error(), "artifactName must be provided") {
		t.Fatalf("artifactName blank: unexpected err %v", err)
	}

	blankPath := ""
	_, err = evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{
		Manifest:     map[string]any{},
		ArtifactPath: &blankPath,
	})
	if err == nil || !strings.Contains(err.Error(), "artifactPath must be provided") {
		t.Fatalf("artifactPath blank: unexpected err %v", err)
	}
}

func TestBuildEvidenceArtifactPlan_defaults_metadata(t *testing.T) {
	t.Parallel()

	manifest := map[string]any{
		"pr_count":      2,
		"pull_requests": []any{},
	}

	plan, err := evidence.BuildEvidenceArtifactPlan(evidence.ArtifactPlanRequest{Manifest: manifest})
	if err != nil {
		t.Fatal(err)
	}

	wantFormatted := strings.TrimSpace(`{
  "pr_count": 2,
  "pull_requests": []
}`)

	gotFormatted := strings.TrimSpace(plan.FormattedJSON)
	if gotFormatted != wantFormatted {
		t.Fatalf("formatted JSON:\n%s\nwant:\n%s", gotFormatted, wantFormatted)
	}

	if plan.ArtifactName != evidence.DefaultEvidenceArtifactName ||
		plan.ArtifactPath != evidence.DefaultEvidenceArtifactPath ||
		plan.Directory != "/tmp/gh-aw/pre-activation" ||
		plan.PRCountDisplay != 2 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestBuildTargetSection_release_and_unreleased(t *testing.T) {
	t.Parallel()

	genAt := time.Date(2026, 4, 17, 8, 0, 0, 0, time.UTC)

	got := evidence.BuildTargetSection("release", "1.2.3", genAt)
	want := "## [1.2.3] - 2026-04-17"
	if got != want {
		t.Fatalf("release: got %q want %q", got, want)
	}

	if got := evidence.BuildTargetSection("unreleased", "1.2.3", genAt); got != "## [Unreleased]" {
		t.Fatalf("unreleased: %q", got)
	}
}
