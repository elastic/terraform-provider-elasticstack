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

// Command changelog implements GitHub workflow helpers for changelog generation
// and PR changelog validation.
//
// Usage:
//
//	go run ./scripts/changelog <subcommand> [flags]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/engine"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "changelog: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError(stderr)
	}
	switch args[0] {
	case "gather-evidence":
		return cmdGatherEvidence(args[1:], stderr)
	case "run-engine":
		return cmdRunEngine(args[1:], stderr)
	case "manage-unreleased-pr":
		return cmdManageUnreleasedPR(args[1:], stderr)
	case "refresh-release-pr":
		return cmdRefreshReleasePR(args[1:], stderr)
	case "validate-pr-section":
		return cmdValidatePRSection(args[1:], stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: changelog <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  gather-evidence        Build evidence manifest for a release range")
	fmt.Fprintln(w, "  run-engine             Run unreleased or release changelog engine")
	fmt.Fprintln(w, "  manage-unreleased-pr   Open or update the generated-changelog PR")
	fmt.Fprintln(w, "  refresh-release-pr     Refresh the release changelog PR")
	fmt.Fprintln(w, "  validate-pr-section    Validate PR body ## Changelog section")
	return errors.New("unknown or missing subcommand")
}

func cmdGatherEvidence(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("gather-evidence", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: gather-evidence")
}

func cmdRunEngine(args []string, stderr io.Writer) error {
	ctx := context.Background()
	fsFlag := flag.NewFlagSet("run-engine", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	modeFlag := fsFlag.String("mode", "",
		"changelog mode: unreleased|release (defaults to $MODE or unreleased)")
	targetVerFlag := fsFlag.String("target-version", "",
		"semver X.Y.Z without leading v for release (defaults to $TARGET_VERSION)")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	mode := strings.TrimSpace(*modeFlag)
	if mode == "" {
		mode = strings.TrimSpace(os.Getenv(githubx.EnvMode))
	}
	if mode == "" {
		mode = engine.ModeUnreleased
	}

	targetVersion := strings.TrimSpace(*targetVerFlag)
	if targetVersion == "" {
		targetVersion = strings.TrimSpace(os.Getenv(githubx.EnvTargetVersion))
	}

	changelogPath := strings.TrimSpace(os.Getenv(githubx.EnvChangelogPath))
	if changelogPath == "" {
		changelogPath = "CHANGELOG.md"
	}

	targetBranchOverride := strings.TrimSpace(os.Getenv(githubx.EnvTargetBranch))

	owner, repo, err := githubx.OwnerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("run-engine: github repository env: %w", err)
	}

	client, err := githubx.NewGitHubClient(ctx, githubx.GitHubToken())
	if err != nil {
		return err
	}

	gather := &gitMergedPRGatherer{client: client, execer: githubx.ShellGit{}}

	res, err := engine.Run(ctx, engine.Options{
		Mode:          mode,
		TargetVersion: targetVersion,
		Owner:         owner,
		Repo:          repo,
		ChangelogPath: changelogPath,
		Now:           time.Now,
		FS:            osChangelogFS{},
		Git:           githubx.ShellGit{},
		Gather:        gather,
	})
	if err != nil {
		return err
	}

	effectiveBranch := targetBranchOverride
	if effectiveBranch == "" {
		effectiveBranch = res.TargetBranch
	}

	outPath := githubx.GitHubOutputPath()
	if outPath != "" {
		appendOut := func(name, value string) error {
			if perr := githubx.AppendGitHubOutput(outPath, name, value); perr != nil {
				return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, perr)
			}
			return nil
		}
		writes := [][2]string{
			{"mode", mode},
			{"target_version", res.TargetVersionOutput},
			{"previous_tag", res.PreviousTag},
			{"compare_range", res.CompareRange},
			{"target_branch", effectiveBranch},
			{"has_prs", boolGitHubActionsString(res.HasPRs)},
			{"has_user_facing_changes", boolGitHubActionsString(res.HasUserFacingChanges)},
			{"section_header", res.SectionHeader},
		}
		for _, kv := range writes {
			if werr := appendOut(kv[0], kv[1]); werr != nil {
				return werr
			}
		}
	}

	return nil
}

// osChangelogFS implements engine.FS backed by OS file operations.
type osChangelogFS struct{}

func (osChangelogFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (osChangelogFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func boolGitHubActionsString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func cmdManageUnreleasedPR(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("manage-unreleased-pr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: manage-unreleased-pr")
}

func cmdRefreshReleasePR(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("refresh-release-pr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: refresh-release-pr")
}

func cmdValidatePRSection(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate-pr-section", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: validate-pr-section")
}
