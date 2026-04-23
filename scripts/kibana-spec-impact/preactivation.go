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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const defaultIssueCap = 5

type PreActivationOutputs struct {
	ShouldRun           bool
	IssueCap            int
	HighConfidenceCount int
	GateReason          string
}

func derivePreActivationOutputs(report *ImpactReport, issueCap int) PreActivationOutputs {
	if issueCap <= 0 {
		issueCap = defaultIssueCap
	}
	outputs := PreActivationOutputs{
		IssueCap: issueCap,
	}
	if report == nil {
		outputs.GateReason = "missing_report"
		return outputs
	}
	outputs.HighConfidenceCount = len(report.HighConfidence)
	if outputs.HighConfidenceCount > 0 {
		outputs.ShouldRun = true
		outputs.GateReason = "high_confidence_impacts_present"
		return outputs
	}
	if len(report.TransformSchemaHints) > 0 {
		outputs.ShouldRun = true
		outputs.GateReason = "transform_schema_hints_present"
		return outputs
	}
	outputs.GateReason = "no_actionable_impacts"
	return outputs
}

func writeReportFile(path string, report *ImpactReport) error {
	data, err := encodeReport(report)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir for report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func appendGithubOutputs(path string, outputs PreActivationOutputs) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open github output: %w", err)
	}
	defer f.Close()
	for key, value := range map[string]string{
		"should_run":            strconv.FormatBool(outputs.ShouldRun),
		"issue_cap":             strconv.Itoa(outputs.IssueCap),
		"high_confidence_count": strconv.Itoa(outputs.HighConfidenceCount),
		"gate_reason":           outputs.GateReason,
	} {
		if _, err := fmt.Fprintf(f, "%s=%s\n", key, value); err != nil {
			return fmt.Errorf("write github output %s: %w", key, err)
		}
	}
	return nil
}
