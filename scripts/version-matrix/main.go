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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

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
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: version-matrix <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  compute   Compute the desired version list and update the pinned artifact")
	return errors.New("unknown or missing subcommand")
}

const (
	defaultArtifactPath      = ".github/versions/acceptance-test-matrix.json"
	defaultSnapshotURL       = "https://snapshots.elastic.co/latest/master.json"
	defaultElasticRegistry   = "https://docker.elastic.co"
	defaultDockerHubRegistry = "https://registry-1.docker.io"
)

func cmdCompute(args []string, stdout, stderr io.Writer) error {
	fsFlag := flag.NewFlagSet("compute", flag.ContinueOnError)
	fsFlag.SetOutput(stderr)
	artifactPath := fsFlag.String("artifact", defaultArtifactPath, "path to the pinned versions artifact")
	githubURL := fsFlag.String("github-url", "", "optional GitHub API base URL")
	snapshotURL := fsFlag.String("snapshot-url", defaultSnapshotURL, "SNAPSHOT label endpoint")
	elasticRegistry := fsFlag.String("elastic-registry", defaultElasticRegistry, "docker.elastic.co registry base URL")
	dockerHubRegistry := fsFlag.String("dockerhub-registry", defaultDockerHubRegistry, "Docker Hub registry base URL")
	if err := fsFlag.Parse(args); err != nil {
		return err
	}

	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		return errors.New("missing GITHUB_TOKEN")
	}

	opts := []github.ClientOptionsFunc{github.WithAuthToken(token)}
	if strings.TrimSpace(*githubURL) != "" {
		opts = append(opts, github.WithEnterpriseURLs(*githubURL, *githubURL))
	}
	client, err := github.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("github client: %w", err)
	}

	ctx := context.Background()
	tags, err := ListElasticsearchTags(ctx, client)
	if err != nil {
		return err
	}

	snapshot, err := FetchSnapshotLabel(ctx, http.DefaultClient, *snapshotURL)
	if err != nil {
		return err
	}

	pinned, err := readPinnedArtifact(*artifactPath)
	if err != nil {
		return err
	}

	prober := RegistryProber{
		Client:            http.DefaultClient,
		ElasticRegistry:   *elasticRegistry,
		DockerHubRegistry: *dockerHubRegistry,
	}
	desired := ComputeDesired(tags, snapshot, pinned, func(version string) bool {
		ok, perr := ProbeComposeStack(ctx, prober, version)
		return perr == nil && ok
	})

	if VersionsEqual(desired, pinned) {
		fmt.Fprintln(stdout, "changed=false")
		return nil
	}
	if err := WriteArtifact(*artifactPath, desired); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "changed=true")
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
