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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v85/github"
	"golang.org/x/oauth2"
)

func main() {
	// Support legacy subcommand-based invocation for backward compatibility:
	//   go run ./scripts/changelog-generation <subcommand> [flags]
	//
	// When invoked without arguments (default GitHub Actions mode), the engine
	// reads configuration from environment variables and runs the workflow engine.
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "validate-provenance":
			runValidateProvenance(os.Args[2:])
			return
		case "rewrite-changelog-section":
			runRewriteChangelogSection(os.Args[2:])
			return
		case "build-evidence-manifest":
			runBuildEvidenceManifest(os.Args[2:])
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "changelog-generation error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/changelog-generation <subcommand> [flags]")
	fmt.Fprintln(os.Stderr, "       go run ./scripts/changelog-generation  (no subcommand: GitHub Actions engine mode)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  validate-provenance       Validate provenance JSON against evidence manifest")
	fmt.Fprintln(os.Stderr, "  rewrite-changelog-section Rewrite a section in CHANGELOG.md")
	fmt.Fprintln(os.Stderr, "  build-evidence-manifest   Build an evidence manifest JSON")
}

func run(ctx context.Context) error {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		return errors.New("missing GITHUB_TOKEN")
	}
	repoValue := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY"))
	owner, repo, err := parseRepository(repoValue)
	if err != nil {
		return err
	}
	mode := strings.TrimSpace(os.Getenv("CHANGELOG_MODE"))
	targetVersion := strings.TrimSpace(os.Getenv("TARGET_VERSION"))
	changelogPath := strings.TrimSpace(os.Getenv("CHANGELOG_PATH"))
	if changelogPath == "" {
		changelogPath = "CHANGELOG.md"
	}

	client := githubClient(ctx, token)
	result, err := runChangelogEngine(ctx, client, owner, repo, mode, targetVersion, changelogPath, time.Now().UTC())
	if err != nil {
		return err
	}

	if outputFile := os.Getenv("GITHUB_OUTPUT"); outputFile != "" {
		lines := []string{
			fmt.Sprintf("mode=%s", result.Mode),
			fmt.Sprintf("target_version=%s", result.TargetVersion),
			fmt.Sprintf("target_branch=%s", result.TargetBranch),
			fmt.Sprintf("previous_tag=%s", result.PreviousTag),
			fmt.Sprintf("compare_range=%s", result.CompareRange),
			fmt.Sprintf("section_header=%s", result.SectionHeader),
			fmt.Sprintf("has_user_facing_changes=%t", result.HasUserFacingChanges),
			fmt.Sprintf("has_pull_requests=%t", len(result.PullRequests) > 0),
		}
		for _, line := range lines {
			if err := appendToFile(outputFile, line+"\n"); err != nil {
				return fmt.Errorf("write GITHUB_OUTPUT: %w", err)
			}
		}
	}

	fmt.Printf("generated changelog section %q from %s\n", result.SectionHeader, result.CompareRange)
	return nil
}

func githubClient(ctx context.Context, token string) *github.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, tokenSource)
	return github.NewClient(httpClient)
}

func parseRepository(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid GITHUB_REPOSITORY: %q", repo)
	}
	return parts[0], parts[1], nil
}

func appendToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
