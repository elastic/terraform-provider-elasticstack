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

// Command version-matrix computes the acceptance-test stack version list.
//
// Usage:
//
//	go run ./scripts/version-matrix <subcommand> [flags]
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v89/github"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "version-matrix: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError(stderr)
	}
	switch args[0] {
	case "compute":
		return cmdCompute(args[1:], stdout, stderr)
	case "manage-pr":
		return cmdManagePR(args[1:], stdout, stderr)
	case "load-matrix":
		return cmdLoadMatrix(args[1:], stdout, stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: version-matrix <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  compute     Compute the desired version list and update the pinned artifact")
	fmt.Fprintln(w, "  manage-pr   Create or update the standing version-matrix pull request")
	fmt.Fprintln(w, "  load-matrix Read the pinned artifact and emit matrix versions and flags")
	return errors.New("unknown or missing subcommand")
}

func cmdManagePR(args []string, stdout, stderr io.Writer) error {
	fsFlag := flag.NewFlagSet("manage-pr", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	changed := fsFlag.Bool("changed", true, "whether the computed version list differs from the pin")
	githubURL := fsFlag.String("github-url", "", "optional GitHub API base URL")
	baseBranch := fsFlag.String("base-branch", "", "PR base branch (defaults to $DEFAULT_BRANCH or main)")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	if !*changed {
		_, err := ManageStandingPR(context.Background(), ManageStandingOptions{Changed: false})
		return err
	}

	owner, repo, err := ownerRepoFromEnv()
	if err != nil {
		return fmt.Errorf("manage-pr: %w", err)
	}
	client, err := newGitHubClient(strings.TrimSpace(os.Getenv("GITHUB_TOKEN")), *githubURL)
	if err != nil {
		return fmt.Errorf("manage-pr: %w", err)
	}

	base := strings.TrimSpace(*baseBranch)
	if base == "" {
		base = strings.TrimSpace(os.Getenv("DEFAULT_BRANCH"))
	}

	res, err := ManageStandingPR(context.Background(), ManageStandingOptions{
		Changed:    true,
		Owner:      owner,
		Repo:       repo,
		BaseBranch: base,
		GitHub:     &githubStandingREST{client: client},
		Now:        time.Now,
	})
	if err != nil {
		return fmt.Errorf("manage-pr: %w", err)
	}
	for _, w := range res.Warnings {
		writeWorkflowWarning(stdout, w)
	}
	writes := [][2]string{
		{"pr_action", res.Action},
		{"pr_number", fmt.Sprintf("%d", res.Number)},
		{"pr_url", res.URL},
	}
	for _, kv := range writes {
		if werr := writeGitHubOutput(kv[0], kv[1]); werr != nil {
			return fmt.Errorf("manage-pr: %w", werr)
		}
	}
	return nil
}

func writeWorkflowWarning(w io.Writer, msg string) {
	fmt.Fprintf(w, "::warning::%s\n", escapeGitHubWorkflowData(msg))
}

func escapeGitHubWorkflowData(msg string) string {
	msg = strings.ReplaceAll(msg, "%", "%25")
	msg = strings.ReplaceAll(msg, "\r", "%0D")
	msg = strings.ReplaceAll(msg, "\n", "%0A")
	return msg
}

func ownerRepoFromEnv() (owner, repo string, err error) {
	raw := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY"))
	parts := strings.Split(raw, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GITHUB_REPOSITORY value %q", raw)
	}
	return parts[0], parts[1], nil
}

// newGitHubClient builds a github.Client authenticated with token, optionally
// pointed at an enterprise API base URL, with any extra client options applied.
func newGitHubClient(token, enterpriseURL string, extra ...github.ClientOptionsFunc) (*github.Client, error) {
	if token == "" {
		return nil, errors.New("missing GITHUB_TOKEN")
	}
	opts := append([]github.ClientOptionsFunc{github.WithAuthToken(token)}, extra...)
	if strings.TrimSpace(enterpriseURL) != "" {
		opts = append(opts, github.WithEnterpriseURLs(enterpriseURL, enterpriseURL))
	}
	return github.NewClient(opts...)
}

const (
	defaultArtifactPath      = ".github/versions/acceptance-test-matrix.json"
	defaultSnapshotURL       = "https://snapshots.elastic.co/latest/master.json"
	defaultElasticRegistry   = "https://docker.elastic.co"
	defaultDockerHubRegistry = "https://registry-1.docker.io"
	defaultComputeTimeout    = 2 * time.Minute
)

func newComputeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func cmdCompute(args []string, stdout, stderr io.Writer) error {
	fsFlag := flag.NewFlagSet("compute", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	artifactPath := fsFlag.String("artifact", defaultArtifactPath, "path to the pinned versions artifact")
	githubURL := fsFlag.String("github-url", "", "optional GitHub API base URL")
	snapshotURL := fsFlag.String("snapshot-url", defaultSnapshotURL, "SNAPSHOT label endpoint")
	elasticRegistry := fsFlag.String("elastic-registry", defaultElasticRegistry, "docker.elastic.co registry base URL")
	dockerHubRegistry := fsFlag.String("dockerhub-registry", defaultDockerHubRegistry, "Docker Hub registry base URL")
	timeout := fsFlag.Duration("timeout", defaultComputeTimeout, "maximum time for the compute run")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	httpClient := newComputeHTTPClient(*timeout)
	client, err := newGitHubClient(strings.TrimSpace(os.Getenv("GITHUB_TOKEN")), *githubURL, github.WithHTTPClient(httpClient))
	if err != nil {
		return fmt.Errorf("compute: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	tags, err := ListElasticsearchTags(ctx, client)
	if err != nil {
		return err
	}

	snapshot, err := FetchSnapshotLabel(ctx, httpClient, *snapshotURL)
	if err != nil {
		return err
	}

	pinned, err := readPinnedArtifact(*artifactPath)
	if err != nil {
		return err
	}

	prober := RegistryProber{
		Client:            httpClient,
		ElasticRegistry:   *elasticRegistry,
		DockerHubRegistry: *dockerHubRegistry,
	}
	desired, err := ComputeDesired(tags, snapshot, pinned, func(version string) (bool, error) {
		return ProbeComposeStack(ctx, prober, version)
	})
	if err != nil {
		return err
	}

	if VersionsEqual(desired, pinned) {
		fmt.Fprintln(stdout, "changed=false")
		return writeGitHubOutput("changed", "false")
	}
	if err := WriteArtifact(*artifactPath, desired); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "changed=true")
	return writeGitHubOutput("changed", "true")
}

func cmdLoadMatrix(args []string, _, stderr io.Writer) error {
	fsFlag := flag.NewFlagSet("load-matrix", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	artifactPath := fsFlag.String("artifact", defaultArtifactPath, "path to the pinned versions artifact")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	versions, err := ReadArtifact(*artifactPath)
	if err != nil {
		return fmt.Errorf("load-matrix: %w", err)
	}
	flags := FlagsByVersion(versions)
	if err := writeGitHubOutputJSON("versions", versions); err != nil {
		return fmt.Errorf("load-matrix: %w", err)
	}
	if err := writeGitHubOutputJSON("flags", flags); err != nil {
		return fmt.Errorf("load-matrix: %w", err)
	}
	return nil
}

func writeGitHubOutputJSON(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, err)
	}
	return writeGitHubOutput(name, string(data))
}

func writeGitHubOutput(name, value string) error {
	path := strings.TrimSpace(os.Getenv("GITHUB_OUTPUT"))
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, err)
	}
	_, writeErr := fmt.Fprintf(f, "%s=%s\n", name, value)
	if closeErr := f.Close(); closeErr != nil && writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		return fmt.Errorf("GITHUB_OUTPUT (%s): %w", name, writeErr)
	}
	return nil
}

func readPinnedArtifact(path string) ([]string, error) {
	pinned, err := ReadArtifact(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return pinned, nil
}
