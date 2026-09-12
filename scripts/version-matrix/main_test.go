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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRequiresSubcommand(t *testing.T) {
	t.Parallel()

	err := run(nil, io.Discard, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subcommand")
}

func TestWriteWorkflowWarning_escapesPercentAndNewline(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	writeWorkflowWarning(&buf, "failed 100%\nretry")
	assert.Equal(t, "::warning::failed 100%25%0Aretry\n", buf.String())
}

func TestRunManagePRNoOpWhenUnchanged(t *testing.T) {
	err := run([]string{"manage-pr", "-changed=false"}, io.Discard, io.Discard)
	require.NoError(t, err)
}

func TestRunManagePRIsAKnownSubcommand(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("GITHUB_TOKEN", "")
	err := run([]string{"manage-pr"}, io.Discard, io.Discard)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unknown or missing subcommand")
}

func TestRunManagePRCreatesWhenNoExistingPR(t *testing.T) {
	srv := startStandingPRServer(t, nil)
	outPath := filepath.Join(t.TempDir(), "github-output")
	t.Setenv("GITHUB_REPOSITORY", "org/repo")
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_OUTPUT", outPath)

	err := run([]string{"manage-pr", "-changed=true", "-github-url", srv.URL}, io.Discard, io.Discard)
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), "pr_action=created")
	assert.Contains(t, string(got), "pr_number=7")
	assert.Contains(t, string(got), "pr_url=https://github.com/org/repo/pull/7")
}

func TestRunManagePRUpdatesWhenExistingPR(t *testing.T) {
	srv := startStandingPRServer(t, []byte(`[{"number":42,"html_url":"https://github.com/org/repo/pull/42"}]`))
	outPath := filepath.Join(t.TempDir(), "github-output")
	t.Setenv("GITHUB_REPOSITORY", "org/repo")
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_OUTPUT", outPath)

	err := run([]string{"manage-pr", "-changed=true", "-github-url", srv.URL}, io.Discard, io.Discard)
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), "pr_action=updated")
	assert.Contains(t, string(got), "pr_number=42")
	assert.Contains(t, string(got), "pr_url=https://github.com/org/repo/pull/42")
}

func startStandingPRServer(t *testing.T, existing []byte) *httptest.Server {
	t.Helper()
	if existing == nil {
		existing = []byte("[]")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/repos/org/repo/pulls"):
			_, _ = w.Write(existing)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/repos/org/repo/pulls"):
			_, _ = w.Write([]byte(`{"number":7,"html_url":"https://github.com/org/repo/pull/7"}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/issues/") && strings.HasSuffix(r.URL.Path, "/labels"):
			_, _ = w.Write([]byte(`[{"name":"no-changelog"}]`))
		case r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/repos/org/repo/pulls/"):
			_, _ = w.Write([]byte(`{"number":42,"html_url":"https://github.com/org/repo/pull/42"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRunComputeUpdatesArtifactWhenChanged(t *testing.T) {
	fx := startComputeFixtures(t, `[{"name":"v8.19.17"},{"name":"v8.19.21"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.17"}))

	t.Setenv("GITHUB_TOKEN", "test-token")

	var stdout bytes.Buffer
	err := run(computeArgs(artifact, fx), &stdout, io.Discard)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "changed=true")

	got, err := ReadArtifact(artifact)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.19.21", "9.6.0-SNAPSHOT"}, got)
}

func TestRunComputeWritesGitHubOutputWhenChanged(t *testing.T) {
	fx := startComputeFixtures(t, `[{"name":"v8.19.17"},{"name":"v8.19.21"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.17"}))

	outPath := filepath.Join(t.TempDir(), "github-output")
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_OUTPUT", outPath)

	err := run(computeArgs(artifact, fx), io.Discard, io.Discard)
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), "changed=true")
}

func TestRunComputeNoOpWhenArtifactMatches(t *testing.T) {
	fx := startComputeFixtures(t, `[{"name":"v8.19.21"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.21", "9.6.0-SNAPSHOT"}))
	before, err := os.ReadFile(artifact)
	require.NoError(t, err)

	t.Setenv("GITHUB_TOKEN", "test-token")

	var stdout bytes.Buffer
	err = run(computeArgs(artifact, fx), &stdout, io.Discard)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "changed=false")

	after, err := os.ReadFile(artifact)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestRunComputeWritesGitHubOutputWhenUnchanged(t *testing.T) {
	fx := startComputeFixtures(t, `[{"name":"v8.19.21"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.21", "9.6.0-SNAPSHOT"}))

	outPath := filepath.Join(t.TempDir(), "github-output")
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_OUTPUT", outPath)

	err := run(computeArgs(artifact, fx), io.Discard, io.Discard)
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), "changed=false")
}

func TestRunComputeFailsWhenSnapshotEndpointUnavailable(t *testing.T) {
	fx := startComputeFixtures(t, `[{"name":"v8.19.21"}]`, http.StatusServiceUnavailable, "")
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.21", "9.6.0-SNAPSHOT"}))
	before, err := os.ReadFile(artifact)
	require.NoError(t, err)

	t.Setenv("GITHUB_TOKEN", "test-token")

	err = run(computeArgs(artifact, fx), io.Discard, io.Discard)
	require.Error(t, err)

	after, err := os.ReadFile(artifact)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestRunComputeFailsWhenImageProbeIsUncertain(t *testing.T) {
	githubSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/repos/elastic/elasticsearch/tags" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"name":"v8.19.21"}]`))
	}))
	t.Cleanup(githubSrv.Close)

	snapshotSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"version":"9.6.0-SNAPSHOT"}`))
	}))
	t.Cleanup(snapshotSrv.Close)

	registrySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(registrySrv.Close)

	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.17"}))
	before, err := os.ReadFile(artifact)
	require.NoError(t, err)

	t.Setenv("GITHUB_TOKEN", "test-token")

	err = run([]string{
		"compute",
		"-artifact", artifact,
		"-github-url", githubSrv.URL,
		"-snapshot-url", snapshotSrv.URL,
		"-elastic-registry", registrySrv.URL,
		"-dockerhub-registry", registrySrv.URL,
	}, io.Discard, io.Discard)
	require.Error(t, err)

	after, err := os.ReadFile(artifact)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestRunComputeFailsWhenGitHubTagListFails(t *testing.T) {
	githubSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(githubSrv.Close)

	fx := startComputeFixtures(t, `[{"name":"v8.19.21"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.21", "9.6.0-SNAPSHOT"}))
	before, err := os.ReadFile(artifact)
	require.NoError(t, err)

	t.Setenv("GITHUB_TOKEN", "test-token")

	err = run([]string{
		"compute",
		"-artifact", artifact,
		"-github-url", githubSrv.URL,
		"-snapshot-url", fx.snapshotURL,
		"-elastic-registry", fx.registryURL,
		"-dockerhub-registry", fx.registryURL,
	}, io.Discard, io.Discard)
	require.Error(t, err)

	after, err := os.ReadFile(artifact)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestComputeHTTPClientHasTimeout(t *testing.T) {
	t.Parallel()

	client := newComputeHTTPClient(30 * time.Second)
	assert.Equal(t, 30*time.Second, client.Timeout)
}

func TestRunComputeFailsWhenHTTPTimesOut(t *testing.T) {
	hang := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(hang.Close)

	fx := startComputeFixtures(t, `[{"name":"v8.19.21"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.21"}))
	before, err := os.ReadFile(artifact)
	require.NoError(t, err)

	t.Setenv("GITHUB_TOKEN", "test-token")

	err = run([]string{
		"compute",
		"-artifact", artifact,
		"-github-url", hang.URL,
		"-snapshot-url", fx.snapshotURL,
		"-elastic-registry", fx.registryURL,
		"-dockerhub-registry", fx.registryURL,
		"-timeout", "50ms",
	}, io.Discard, io.Discard)
	require.Error(t, err)

	after, err := os.ReadFile(artifact)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestRunComputeFailsWhenGATagListEmpty(t *testing.T) {
	fx := startComputeFixtures(t, `[{"name":"v7.17.28"}]`, http.StatusOK, `{"version":"9.6.0-SNAPSHOT"}`)
	artifact := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.19.21", "9.6.0-SNAPSHOT"}))
	before, err := os.ReadFile(artifact)
	require.NoError(t, err)

	t.Setenv("GITHUB_TOKEN", "test-token")

	err = run(computeArgs(artifact, fx), io.Discard, io.Discard)
	require.Error(t, err)

	after, err := os.ReadFile(artifact)
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

type computeFixtureURLs struct {
	githubURL   string
	snapshotURL string
	registryURL string
}

func startComputeFixtures(t *testing.T, tagsJSON string, snapshotStatus int, snapshotBody string) computeFixtureURLs {
	t.Helper()

	githubSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/repos/elastic/elasticsearch/tags" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(tagsJSON))
	}))
	t.Cleanup(githubSrv.Close)

	snapshotSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(snapshotStatus)
		if snapshotBody != "" {
			_, _ = w.Write([]byte(snapshotBody))
		}
	}))
	t.Cleanup(snapshotSrv.Close)

	registrySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(registrySrv.Close)

	return computeFixtureURLs{
		githubURL:   githubSrv.URL,
		snapshotURL: snapshotSrv.URL,
		registryURL: registrySrv.URL,
	}
}

func computeArgs(artifact string, fx computeFixtureURLs) []string {
	return []string{
		"compute",
		"-artifact", artifact,
		"-github-url", fx.githubURL,
		"-snapshot-url", fx.snapshotURL,
		"-elastic-registry", fx.registryURL,
		"-dockerhub-registry", fx.registryURL,
	}
}
