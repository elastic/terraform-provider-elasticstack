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
	"flag"
	"fmt"
	"os"
	"strings"

	changelogengine "github.com/elastic/terraform-provider-elasticstack/scripts/changelog-engine"
)

var runEngine = func(ctx context.Context, cfg changelogengine.Config) (*changelogengine.RunResult, error) {
	engine, err := changelogengine.New(cfg)
	if err != nil {
		return nil, err
	}
	return engine.Run(ctx)
}

func main() {
	var mode string
	var targetVersion string
	var owner string
	var repo string
	var token string
	var changelogPath string
	var outputPath string

	flag.StringVar(&mode, "mode", envOrDefault("CHANGELOG_MODE", "unreleased"), "unreleased or release")
	flag.StringVar(&targetVersion, "target-version", os.Getenv("CHANGELOG_TARGET_VERSION"), "target version for release mode")
	flag.StringVar(&owner, "owner", os.Getenv("GITHUB_REPOSITORY_OWNER"), "GitHub owner")
	repoDefault := ""
	if repository := os.Getenv("GITHUB_REPOSITORY"); repository != "" {
		parts := strings.SplitN(repository, "/", 2)
		if len(parts) == 2 {
			repoDefault = parts[1]
		}
	}
	flag.StringVar(&repo, "repo", repoDefault, "GitHub repo")
	flag.StringVar(&token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub token")
	flag.StringVar(&changelogPath, "changelog", envOrDefault("CHANGELOG_PATH", "CHANGELOG.md"), "path to changelog")
	flag.StringVar(&outputPath, "output", "", "optional JSON output file")
	flag.Parse()

	if _, err := run(context.Background(), changelogengine.Config{
		Mode:          changelogengine.Mode(mode),
		TargetVersion: targetVersion,
		Owner:         owner,
		Repo:          repo,
		Token:         token,
		ChangelogPath: changelogPath,
	}, outputPath); err != nil {
		fatal(err)
	}
}

func run(ctx context.Context, cfg changelogengine.Config, outputPath string) (*changelogengine.RunResult, error) {
	result, err := runEngine(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if outputPath != "" {
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(outputPath, payload, 0o644); err != nil {
			return nil, err
		}
	}

	if err := printGitHubOutput(result.Outputs); err != nil {
		return nil, err
	}

	return result, nil
}

func printGitHubOutput(outputs changelogengine.Outputs) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "mode=%s\n", outputs.Mode)
	fmt.Fprintf(f, "target_version=%s\n", outputs.TargetVersion)
	fmt.Fprintf(f, "target_branch=%s\n", outputs.TargetBranch)
	fmt.Fprintf(f, "previous_tag=%s\n", outputs.PreviousTag)
	fmt.Fprintf(f, "compare_range=%s\n", outputs.CompareRange)
	fmt.Fprintf(f, "section_header=%s\n", outputs.SectionHeader)
	fmt.Fprintf(f, "has_changes=%t\n", outputs.HasChanges)
	fmt.Fprintf(f, "has_user_facing_changes=%t\n", outputs.HasUserFacingChanges)
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
