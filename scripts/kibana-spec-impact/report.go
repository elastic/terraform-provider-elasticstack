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
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// ImpactReport is the deterministic evidence contract consumed by the workflow agent.
type ImpactReport struct {
	Version              int                `json:"version"`
	BaselineSHA          string             `json:"baseline_sha"`
	TargetSHA            string             `json:"target_sha"`
	ChangedKbapiSymbols  []string           `json:"changed_kbapi_symbols"`
	TransformSchemaHints []string           `json:"transform_schema_hints"`
	HighConfidence       []ImpactedEntity   `json:"high_confidence_impacts"`
	SuppressedDuplicates []SuppressedImpact `json:"suppressed_duplicates"`
}

// ImpactedEntity is a deterministic high-confidence match for a single Terraform entity.
type ImpactedEntity struct {
	EntityType     string   `json:"entity_type"`
	EntityName     string   `json:"entity_name"`
	PkgPath        string   `json:"pkg_path"`
	MatchedSymbols []string `json:"matched_symbols"`
	Confidence     string   `json:"confidence"`
	Fingerprint    string   `json:"fingerprint"`
}

// SuppressedImpact records an impact that was already reported for this baseline→target identity.
type SuppressedImpact struct {
	EntityName  string `json:"entity_name"`
	Fingerprint string `json:"fingerprint"`
	Reason      string `json:"reason"`
}

func buildImpactReport(repoRoot string, mem *Memory, baselineSHA, targetSHA string) (*ImpactReport, error) {
	baselineSHA, err := gitRevParse(repoRoot, baselineSHA)
	if err != nil {
		return nil, err
	}
	targetSHA, err = gitRevParse(repoRoot, targetSHA)
	if err != nil {
		return nil, err
	}

	changed, err := diffKbapiAtRefs(repoRoot, baselineSHA, targetSHA)
	if err != nil {
		return nil, fmt.Errorf("diff kbapi: %w", err)
	}

	transformHints, err := diffTransformSchemaPaths(repoRoot, baselineSHA, targetSHA)
	if err != nil {
		return nil, fmt.Errorf("transform schema hints: %w", err)
	}

	entities := discoverKibanaEntities()

	oapi, err := buildKibanaOAPIIndex(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("kibanaoapi index: %w", err)
	}

	report := &ImpactReport{
		Version:              1,
		BaselineSHA:          baselineSHA,
		TargetSHA:            targetSHA,
		ChangedKbapiSymbols:  changed,
		TransformSchemaHints: append([]string{}, transformHints...),
		HighConfidence:       []ImpactedEntity{},
		SuppressedDuplicates: []SuppressedImpact{},
	}

	for _, e := range entities {
		paths, err := entityScanPaths(repoRoot, e)
		if err != nil {
			return nil, err
		}
		if len(paths) == 0 {
			continue
		}
		matched, err := matchHighConfidence(paths, oapi, changed)
		if err != nil {
			return nil, err
		}
		high, suppressed := impactEntryForEntity(mem, baselineSHA, targetSHA, e, matched)
		if suppressed != nil {
			report.SuppressedDuplicates = append(report.SuppressedDuplicates, *suppressed)
			continue
		}
		if high != nil {
			report.HighConfidence = append(report.HighConfidence, *high)
		}
	}

	return report, nil
}

// impactEntryForEntity maps a matched entity to a high-confidence impact or a suppressed duplicate.
func impactEntryForEntity(mem *Memory, baselineSHA, targetSHA string, e Entity, matched []string) (high *ImpactedEntity, suppressed *SuppressedImpact) {
	if len(matched) == 0 {
		return nil, nil
	}
	fp := impactFingerprint(baselineSHA, targetSHA, e.Name, e.Type, matched)
	if mem != nil && memoryIsReported(mem, fp) {
		return nil, &SuppressedImpact{
			EntityName:  e.Name,
			Fingerprint: fp,
			Reason:      "duplicate_fingerprint",
		}
	}
	return &ImpactedEntity{
		EntityType:     e.Type,
		EntityName:     e.Name,
		PkgPath:        e.PkgPath,
		MatchedSymbols: matched,
		Confidence:     "high",
		Fingerprint:    fp,
	}, nil
}

func diffTransformSchemaPaths(repoRoot, baselineSHA, targetSHA string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "diff", "--name-only", baselineSHA+".."+targetSHA)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff name-only: %w", err)
	}
	var paths []string
	for line := range strings.SplitSeq(strings.TrimSpace(stdout.String()), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "internal/kibana/") {
			continue
		}
		if !strings.HasSuffix(line, "transform_schema.go") {
			continue
		}
		paths = append(paths, filepath.ToSlash(line))
	}
	sort.Strings(paths)
	return paths, nil
}

func encodeReport(r *ImpactReport) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
