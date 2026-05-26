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
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
)

type curatedFeatures struct {
	Documented []string `json:"documented"`
	Skip       []string `json:"skip"`
}

type driftReport struct {
	UnknownFeatures []string `json:"unknown_features"`
	RemovedFeatures []string `json:"removed_features"`
}

type apiFeature struct {
	ID string `json:"id"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected subcommand")
	}

	switch args[0] {
	case "pre-activation":
		return runPreActivation(args[1:], stdout)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runPreActivation(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("pre-activation", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	featuresPath := fs.String("features-path", "scripts/security-role-docs/kibana-features.json", "path to curated Kibana features JSON")
	reportPath := fs.String("report-path", "", "path to write the drift report JSON")
	issueCap := fs.Int("issue-cap", 1, "reserved for workflow compatibility")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_ = issueCap

	if strings.TrimSpace(*reportPath) == "" {
		return errors.New("--report-path is required")
	}

	features, err := loadCuratedFeatures(*featuresPath)
	if err != nil {
		return err
	}

	apiFeatures, err := fetchFeatureIDs(context.Background())
	if err != nil {
		return err
	}

	report := computeDrift(features, apiFeatures)
	if err := writeJSON(*reportPath, report); err != nil {
		return err
	}

	runAgent := len(report.UnknownFeatures) > 0 || len(report.RemovedFeatures) > 0
	if err := writeGithubOutput("run_agent", fmt.Sprintf("%t", runAgent)); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "run_agent=%t\n", runAgent)
	return err
}

func fetchFeatureIDs(ctx context.Context) ([]string, error) {
	cfg := kibanaoapi.Config{
		URL:      strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT")),
		Username: strings.TrimSpace(os.Getenv("KIBANA_USERNAME")),
		Password: os.Getenv("KIBANA_PASSWORD"),
	}
	if cfg.URL == "" {
		return nil, errors.New("KIBANA_ENDPOINT must be set")
	}

	client, err := kibanaoapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(client.URL, "/")+"/api/features", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/features returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var features []apiFeature
	if err := json.Unmarshal(body, &features); err != nil {
		return nil, fmt.Errorf("parse /api/features response: %w", err)
	}

	ids := make([]string, 0, len(features))
	for _, feature := range features {
		if strings.TrimSpace(feature.ID) == "" {
			continue
		}
		ids = append(ids, feature.ID)
	}
	return sortedUnique(ids), nil
}

func loadCuratedFeatures(path string) (curatedFeatures, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return curatedFeatures{}, fmt.Errorf("read features file: %w", err)
	}

	var features curatedFeatures
	if err := json.Unmarshal(content, &features); err != nil {
		return curatedFeatures{}, fmt.Errorf("parse features file: %w", err)
	}

	features.Documented = sortedUnique(features.Documented)
	features.Skip = sortedUnique(features.Skip)
	return features, nil
}

func computeDrift(features curatedFeatures, apiFeatures []string) driftReport {
	documented := make(map[string]struct{}, len(features.Documented))
	for _, feature := range features.Documented {
		documented[feature] = struct{}{}
	}

	skipped := make(map[string]struct{}, len(features.Skip))
	for _, feature := range features.Skip {
		skipped[feature] = struct{}{}
	}

	apiSet := make(map[string]struct{}, len(apiFeatures))
	unknown := make([]string, 0)
	for _, feature := range sortedUnique(apiFeatures) {
		apiSet[feature] = struct{}{}
		if _, ok := documented[feature]; ok {
			continue
		}
		if _, ok := skipped[feature]; ok {
			continue
		}
		unknown = append(unknown, feature)
	}

	removed := make([]string, 0)
	for _, feature := range features.Documented {
		if _, ok := apiSet[feature]; !ok {
			removed = append(removed, feature)
		}
	}

	return driftReport{
		UnknownFeatures: sortedUnique(unknown),
		RemovedFeatures: sortedUnique(removed),
	}
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func writeGithubOutput(name, value string) error {
	githubOutput := strings.TrimSpace(os.Getenv("GITHUB_OUTPUT"))
	if githubOutput == "" {
		return nil
	}

	f, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s=%s\n", name, value)
	return err
}

func sortedUnique(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}
