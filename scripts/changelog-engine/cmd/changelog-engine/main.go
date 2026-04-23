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
	flag.StringVar(&outputPath, "output", os.Getenv("GITHUB_OUTPUT"), "optional JSON output file")
	flag.Parse()

	engine, err := changelogengine.New(changelogengine.Config{
		Mode:          changelogengine.Mode(mode),
		TargetVersion: targetVersion,
		Owner:         owner,
		Repo:          repo,
		Token:         token,
		ChangelogPath: changelogPath,
	})
	if err != nil {
		fatal(err)
	}

	result, err := engine.Run(context.Background())
	if err != nil {
		fatal(err)
	}

	if outputPath != "" {
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fatal(err)
		}
		if err := os.WriteFile(outputPath, payload, 0o644); err != nil {
			fatal(err)
		}
	}

	printGitHubOutput(result.Outputs)
}

func printGitHubOutput(outputs changelogengine.Outputs) {
	if path := os.Getenv("GITHUB_OUTPUT"); path != "" {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fatal(err)
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
	}
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
