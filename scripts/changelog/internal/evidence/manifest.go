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

// Package evidence builds PR evidence manifest JSON and the release artifact plan
// (parity with the prior JavaScript evidence manifest helpers).
package evidence

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// JS default artifact metadata from the legacy manifest helper (exported for callers/tests).
const (
	DefaultEvidenceArtifactName = "changelog-release-evidence"
	// DefaultEvidenceArtifactPath mirrors DEFAULT_EVIDENCE_ARTIFACT_PATH in JS.
	DefaultEvidenceArtifactPath = "/tmp/gh-aw/pre-activation/evidence.json"

	modeRelease     = "release"
	modeUnreleased  = "unreleased"
	releaseHeadLine = "## [Unreleased]"
)

const (
	classUserFacing = "user-facing"
	classInternal   = "internal"
	classUncertain  = "uncertain"
)

// PullRequestEvidence mirrors buildPullRequestEvidence output in changelog-pr-evidence.js.
// JSON field names MUST stay aligned with the JS keys.
type PullRequestEvidence struct {
	Number             int      `json:"number"`
	Title              string   `json:"title"`
	URL                string   `json:"url"`
	MergeCommitSHA     string   `json:"merge_commit_sha"`
	Author             string   `json:"author"`
	Labels             []string `json:"labels"`
	TouchedFiles       []string `json:"touched_files"`
	Classification     string   `json:"classification"`
	InclusionRationale *string  `json:"inclusion_rationale"`
	ExclusionRationale *string  `json:"exclusion_rationale"`
}

// Manifest mirrors buildEvidenceManifest JSON in changelog-pr-evidence.js.
// Struct field order affects json.MarshalIndent key ordering — keep aligned with JS.
type Manifest struct {
	GeneratedAt       string                `json:"generated_at"`
	Mode              string                `json:"mode"`
	TargetSection     string                `json:"target_section"`
	TargetSectionMode string                `json:"target_section_mode"`
	TargetVersion     string                `json:"target_version"`
	PreviousTag       string                `json:"previous_tag"`
	CompareRange      string                `json:"compare_range"`
	PRCount           int                   `json:"pr_count"`
	UserFacingCount   int                   `json:"user_facing_count"`
	InternalCount     int                   `json:"internal_count"`
	UncertainCount    int                   `json:"uncertain_count"`
	PullRequests      []PullRequestEvidence `json:"pull_requests"`
}

// ArtifactPlan mirrors buildEvidenceArtifactPlan in changelog-evidence-manifest.js.
type ArtifactPlan struct {
	ArtifactName   string
	ArtifactPath   string
	Directory      string
	FormattedJSON  string
	PRCountDisplay any
}

// ArtifactPlanRequest carries inputs for BuildEvidenceArtifactPlan.
// ArtifactName / ArtifactPath: nil selects JS defaults; a non-nil empty string rejects as in JS core.getInput.
type ArtifactPlanRequest struct {
	Manifest     any
	ArtifactName *string
	ArtifactPath *string
}

// FormatGeneratedAtISO matches Date.prototype.toISOString millisecond formatting used in JS manifests.
func FormatGeneratedAtISO(t time.Time) string {
	u := t.UTC()
	return fmt.Sprintf(
		"%04d-%02d-%02dT%02d:%02d:%02d.%03dZ",
		u.Year(), u.Month(), u.Day(), u.Hour(), u.Minute(), u.Second(), u.Nanosecond()/1e6,
	)
}

// BuildTargetSection mirrors buildTargetSection in changelog-pr-evidence.js.
func BuildTargetSection(mode, targetVersion string, generatedAt time.Time) string {
	if mode == modeRelease && strings.TrimSpace(targetVersion) != "" {
		date, _, _ := strings.Cut(FormatGeneratedAtISO(generatedAt), "T")
		return fmt.Sprintf("## [%s] - %s", targetVersion, date)
	}
	return releaseHeadLine
}

func countByClassification(evidence []PullRequestEvidence) (userFacing, internal, uncertain int) {
	for _, e := range evidence {
		switch e.Classification {
		case classUserFacing:
			userFacing++
		case classInternal:
			internal++
		case classUncertain:
			uncertain++
		}
	}
	return userFacing, internal, uncertain
}

// BuildEvidenceManifest mirrors buildEvidenceManifest in changelog-pr-evidence.js.
func BuildEvidenceManifest(
	mode, targetVersion, previousTag, compareRange string,
	evidence []PullRequestEvidence,
	generatedAt time.Time,
) Manifest {
	uf, in, un := countByClassification(evidence)
	modeNorm := strings.TrimSpace(mode)
	if modeNorm == "" {
		modeNorm = modeUnreleased
	}

	prs := evidence
	if prs == nil {
		prs = []PullRequestEvidence{}
	}

	return Manifest{
		GeneratedAt:       FormatGeneratedAtISO(generatedAt),
		Mode:              modeNorm,
		TargetSection:     BuildTargetSection(modeNorm, targetVersion, generatedAt),
		TargetSectionMode: modeNorm,
		TargetVersion:     targetVersion,
		PreviousTag:       previousTag,
		CompareRange:      compareRange,
		PRCount:           len(evidence),
		UserFacingCount:   uf,
		InternalCount:     in,
		UncertainCount:    un,
		PullRequests:      prs,
	}
}

func manifestDisplayPRCount(m Manifest) any {
	// Mirrors manifest.pr_count ?? '?' when using generic objects; typed Manifest always has int.
	return m.PRCount
}

func validateManifestReceiver(manifest any) error {
	if manifest == nil {
		return errors.New("manifest must be a non-null object")
	}

	v := reflect.ValueOf(manifest)
	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			return errors.New("manifest must be a non-null object")
		}
	case reflect.Struct:
	case reflect.Pointer:
		if v.IsNil() {
			return errors.New("manifest must be a non-null object")
		}
		switch v.Elem().Kind() {
		case reflect.Map, reflect.Struct:
		default:
			return errors.New("manifest must be a non-null object")
		}
	default:
		return errors.New("manifest must be a non-null object")
	}
	return nil
}

func resolveArtifactNaming(artifactName, artifactPath *string) (string, string, error) {
	name := DefaultEvidenceArtifactName
	pathVal := DefaultEvidenceArtifactPath

	if artifactName != nil {
		if strings.TrimSpace(*artifactName) == "" {
			return "", "", errors.New("artifactName must be provided")
		}
		name = strings.TrimSpace(*artifactName)
	}

	if artifactPath != nil {
		if strings.TrimSpace(*artifactPath) == "" {
			return "", "", errors.New("artifactPath must be provided")
		}
		pathVal = strings.TrimSpace(*artifactPath)
	}

	return name, pathVal, nil
}

// BuildEvidenceArtifactPlan mirrors buildEvidenceArtifactPlan in changelog-evidence-manifest.js.
func BuildEvidenceArtifactPlan(req ArtifactPlanRequest) (ArtifactPlan, error) {
	if err := validateManifestReceiver(req.Manifest); err != nil {
		return ArtifactPlan{}, err
	}

	name, pathVal, err := resolveArtifactNaming(req.ArtifactName, req.ArtifactPath)
	if err != nil {
		return ArtifactPlan{}, err
	}

	formattedJSON, err := json.MarshalIndent(req.Manifest, "", "  ")
	if err != nil {
		return ArtifactPlan{}, fmt.Errorf("format manifest JSON: %w", err)
	}

	plan := ArtifactPlan{
		ArtifactName:  name,
		ArtifactPath:  pathVal,
		Directory:     filepath.Dir(pathVal),
		FormattedJSON: string(formattedJSON),
	}

	if typed, ok := req.Manifest.(Manifest); ok {
		plan.PRCountDisplay = manifestDisplayPRCount(typed)
		return plan, nil
	}

	if m, ok := req.Manifest.(map[string]any); ok {
		if v, ok := m["pr_count"]; ok {
			plan.PRCountDisplay = v
			return plan, nil
		}
	}

	plan.PRCountDisplay = "?"
	return plan, nil
}
