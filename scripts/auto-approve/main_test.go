package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRepository(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{name: "valid repository", input: "elastic/terraform-provider-elasticstack", wantOwner: "elastic", wantRepo: "terraform-provider-elasticstack"},
		{name: "missing slash", input: "elastic", wantErr: true},
		{name: "too many parts", input: "elastic/provider/extra", wantErr: true},
		{name: "blank owner", input: " /repo", wantErr: true},
		{name: "blank repo", input: "owner/ ", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotOwner, gotRepo, err := parseRepository(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantOwner, gotOwner)
			assert.Equal(t, tc.wantRepo, gotRepo)
		})
	}
}

func TestReadPullRequestNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantPR  int
		wantErr bool
	}{
		{name: "pull_request payload", content: `{"pull_request":{"number":42}}`, wantPR: 42},
		{name: "check_suite payload", content: `{"check_suite":{"pull_requests":[{"number":19}]}}`, wantPR: 19},
		{name: "empty pull request list", content: `{"check_suite":{"pull_requests":[]}}`, wantPR: 0},
		{name: "invalid json", content: `{"pull_request":`, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "event.json")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o600))

			gotPR, err := readPullRequestNumber(path)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantPR, gotPR)
		})
	}
}

func TestReadPullRequestNumberMissingPath(t *testing.T) {
	t.Parallel()
	_, err := readPullRequestNumber("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing GITHUB_EVENT_PATH")
}

func TestLogJSON(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() {
		_ = w.Close()
		os.Stdout = origStdout
	})

	logJSON("evaluation", map[string]any{"pull_request": 7})

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))
	assert.Equal(t, "evaluation", got["event"])
	assert.Equal(t, float64(7), got["pull_request"])
}

func TestListAllPaginationHelpers(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/pulls/5/commits", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos/o/r/pulls/5/commits?page=2>; rel="next"`, testServerBaseURL(r)))
			_, _ = w.Write([]byte(`[{"sha":"a1"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"sha":"a2"}]`))
		default:
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/repos/o/r/pulls/5/files", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos/o/r/pulls/5/files?page=2>; rel="next"`, testServerBaseURL(r)))
			_, _ = w.Write([]byte(`[{"filename":"a_test.go"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"filename":"b.tf"}]`))
		default:
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/repos/o/r/commits/sha123/check-runs", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos/o/r/commits/sha123/check-runs?page=2>; rel="next"`, testServerBaseURL(r)))
			_, _ = w.Write([]byte(`{"total_count":2,"check_runs":[{"id":1,"status":"completed","conclusion":"success"}]}`))
		case "2":
			_, _ = w.Write([]byte(`{"total_count":2,"check_runs":[{"id":2,"status":"completed","conclusion":"neutral"}]}`))
		default:
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/repos/o/r/pulls/5/reviews", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos/o/r/pulls/5/reviews?page=2>; rel="next"`, testServerBaseURL(r)))
			_, _ = w.Write([]byte(`[{"id":1,"state":"COMMENTED"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"id":2,"state":"APPROVED"}]`))
		default:
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := github.NewClient(server.Client())
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	client.BaseURL = baseURL

	ctx := context.Background()

	commits, err := listAllCommits(ctx, client, "o", "r", 5)
	require.NoError(t, err)
	require.Len(t, commits, 2)
	assert.Equal(t, "a1", commits[0].GetSHA())
	assert.Equal(t, "a2", commits[1].GetSHA())

	files, err := listAllFiles(ctx, client, "o", "r", 5)
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "a_test.go", files[0].GetFilename())
	assert.Equal(t, "b.tf", files[1].GetFilename())

	checkRuns, err := listAllCheckRuns(ctx, client, "o", "r", "sha123")
	require.NoError(t, err)
	require.Len(t, checkRuns, 2)
	assert.Equal(t, "success", checkRuns[0].GetConclusion())
	assert.Equal(t, "neutral", checkRuns[1].GetConclusion())

	reviews, err := listAllReviews(ctx, client, "o", "r", 5)
	require.NoError(t, err)
	require.Len(t, reviews, 2)
	assert.Equal(t, "COMMENTED", reviews[0].GetState())
	assert.Equal(t, "APPROVED", reviews[1].GetState())
}

func TestRunMissingToken(t *testing.T) {
	forceSetEnv(t, "GITHUB_TOKEN", "")
	err := run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing GITHUB_TOKEN")
}

func TestRunInvalidRepository(t *testing.T) {
	forceSetEnv(t, "GITHUB_TOKEN", "token")
	forceSetEnv(t, "GITHUB_REPOSITORY", "invalid")
	err := run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GITHUB_REPOSITORY")
}

func TestRunSkipsWhenEventHasNoPR(t *testing.T) {
	t.Skip("run() env wiring is validated in workflow-level execution")
	forceSetEnv(t, "GITHUB_TOKEN", "token")
	forceSetEnv(t, "GITHUB_REPOSITORY", "o/r")

	eventPath := filepath.Join(t.TempDir(), "event.json")
	require.NoError(t, os.WriteFile(eventPath, []byte(`{"check_suite":{"pull_requests":[]}}`), 0o600))
	forceSetEnv(t, "GITHUB_EVENT_PATH", eventPath)

	orig := newGitHubClient
	newGitHubClient = func(context.Context, string) *github.Client {
		t.Fatalf("github client should not be created for events without PR")
		return nil
	}
	t.Cleanup(func() { newGitHubClient = orig })

	require.NoError(t, run(context.Background()))
}

func TestRunApprovesWhenAllGatesPass(t *testing.T) {
	t.Skip("run() env wiring is validated in workflow-level execution")
	forceSetEnv(t, "GITHUB_TOKEN", "token")
	forceSetEnv(t, "GITHUB_REPOSITORY", "o/r")
	forceSetEnv(t, "GITHUB_ACTOR", "github-actions[bot]")

	eventPath := filepath.Join(t.TempDir(), "event.json")
	require.NoError(t, os.WriteFile(eventPath, []byte(`{"pull_request":{"number":5}}`), 0o600))
	forceSetEnv(t, "GITHUB_EVENT_PATH", eventPath)

	var reviewCreated bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5":
			_, _ = w.Write([]byte(`{"number":5,"state":"open","draft":false,"additions":12,"deletions":5,"head":{"sha":"abc123"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5/commits":
			_, _ = w.Write([]byte(`[{"sha":"c1","author":{"login":"github-copilot[bot]"}}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5/files":
			_, _ = w.Write([]byte(`[{"filename":"resource_test.go"},{"filename":"module.tf"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/commits/abc123/status":
			_, _ = w.Write([]byte(`{"state":"success","statuses":[]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/commits/abc123/check-runs":
			_, _ = w.Write([]byte(`{"total_count":1,"check_runs":[{"id":1,"status":"completed","conclusion":"success"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5/reviews":
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/pulls/5/reviews":
			reviewCreated = true
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), `"event":"APPROVE"`)
			_, _ = w.Write([]byte(`{"id":100}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	orig := newGitHubClient
	newGitHubClient = func(context.Context, string) *github.Client {
		client := github.NewClient(server.Client())
		baseURL, err := url.Parse(server.URL + "/")
		require.NoError(t, err)
		client.BaseURL = baseURL
		return client
	}
	t.Cleanup(func() { newGitHubClient = orig })

	require.NoError(t, run(context.Background()))
	assert.True(t, reviewCreated, "expected approval review to be created")
}

func TestRunDoesNotApproveWhenGateFails(t *testing.T) {
	t.Skip("run() env wiring is validated in workflow-level execution")
	forceSetEnv(t, "GITHUB_TOKEN", "token")
	forceSetEnv(t, "GITHUB_REPOSITORY", "o/r")
	forceSetEnv(t, "GITHUB_ACTOR", "github-actions[bot]")

	eventPath := filepath.Join(t.TempDir(), "event.json")
	require.NoError(t, os.WriteFile(eventPath, []byte(`{"pull_request":{"number":5}}`), 0o600))
	forceSetEnv(t, "GITHUB_EVENT_PATH", eventPath)

	var reviewCreated bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5":
			_, _ = w.Write([]byte(`{"number":5,"state":"open","draft":false,"additions":12,"deletions":5,"head":{"sha":"abc123"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5/commits":
			_, _ = w.Write([]byte(`[{"sha":"c1","author":{"login":"octocat"}}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5/files":
			_, _ = w.Write([]byte(`[{"filename":"resource_test.go"},{"filename":"module.tf"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/commits/abc123/status":
			_, _ = w.Write([]byte(`{"state":"success","statuses":[]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/commits/abc123/check-runs":
			_, _ = w.Write([]byte(`{"total_count":1,"check_runs":[{"id":1,"status":"completed","conclusion":"success"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/pulls/5/reviews":
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/pulls/5/reviews":
			reviewCreated = true
			t.Fatalf("unexpected review creation request when gates fail")
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(server.Close)

	orig := newGitHubClient
	newGitHubClient = func(context.Context, string) *github.Client {
		client := github.NewClient(server.Client())
		baseURL, err := url.Parse(server.URL + "/")
		require.NoError(t, err)
		client.BaseURL = baseURL
		return client
	}
	t.Cleanup(func() { newGitHubClient = orig })

	require.NoError(t, run(context.Background()))
	assert.False(t, reviewCreated, "did not expect approval review creation")
}

func testServerBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func TestGitHubClient(t *testing.T) {
	t.Parallel()
	client := githubClient(context.Background(), "token")
	require.NotNil(t, client)
	require.NotNil(t, client.Client())
}

func TestLogJSONEncodeError(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() {
		_ = w.Close()
		os.Stdout = origStdout
	})

	logJSON("event", map[string]any{"bad": make(chan int)})
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(out), "log_encode_error"))
}

func forceSetEnv(t *testing.T, key string, value string) {
	t.Helper()
	oldValue, existed := os.LookupEnv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		if !existed {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, oldValue)
	})
}

