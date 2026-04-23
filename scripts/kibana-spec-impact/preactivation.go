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

const (
	defaultIssueCap                        = 5
	gateReasonMissingReport                = "missing_report"
	gateReasonHighConfidenceImpactsPresent = "high_confidence_impacts_present"
	gateReasonTransformSchemaHintsPresent  = "transform_schema_hints_present"
	gateReasonKbapiSymbolsChanged          = "kbapi_symbols_changed"
	gateReasonNoActionableImpacts          = "no_actionable_impacts"
)

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
		outputs.GateReason = gateReasonMissingReport
		return outputs
	}
	outputs.HighConfidenceCount = len(report.HighConfidence)
	if outputs.HighConfidenceCount > 0 {
		outputs.ShouldRun = true
		outputs.GateReason = gateReasonHighConfidenceImpactsPresent
		return outputs
	}
	if len(report.TransformSchemaHints) > 0 {
		outputs.ShouldRun = true
		outputs.GateReason = gateReasonTransformSchemaHintsPresent
		return outputs
	}
	if len(report.ChangedKbapiSymbols) > 0 {
		outputs.ShouldRun = true
		outputs.GateReason = gateReasonKbapiSymbolsChanged
		return outputs
	}
	outputs.GateReason = gateReasonNoActionableImpacts
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

func appendGithubOutputs(path string, outputs PreActivationOutputs) (err error) {
	f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		return fmt.Errorf("open github output: %w", openErr)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close github output: %w", closeErr)
		}
	}()
	entries := []struct {
		key   string
		value string
	}{
		{key: "run_agent", value: strconv.FormatBool(outputs.ShouldRun)},
		{key: "issue_cap", value: strconv.Itoa(outputs.IssueCap)},
		{key: "high_confidence_count", value: strconv.Itoa(outputs.HighConfidenceCount)},
		{key: "gate_reason", value: outputs.GateReason},
	}
	for _, entry := range entries {
		if _, err := fmt.Fprintf(f, "%s=%s\n", entry.key, entry.value); err != nil {
			return fmt.Errorf("write github output %s: %w", entry.key, err)
		}
	}
	return nil
}
