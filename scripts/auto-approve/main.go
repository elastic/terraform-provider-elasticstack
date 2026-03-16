package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

type eventPayload struct {
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
	CheckSuite *struct {
		PullRequests []struct {
			Number int `json:"number"`
		} `json:"pull_requests"`
	} `json:"check_suite"`
}

var newGitHubClient = githubClient

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "auto-approve error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		return errors.New("missing GITHUB_TOKEN")
	}

	repo := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY"))
	owner, name, err := parseRepository(repo)
	if err != nil {
		return err
	}

	prNumber, err := readPullRequestNumber(strings.TrimSpace(os.Getenv("GITHUB_EVENT_PATH")))
	if err != nil {
		return err
	}
	if prNumber == 0 {
		logJSON("skip", map[string]any{
			"reason": "event has no associated pull request",
		})
		return nil
	}

	client := newGitHubClient(ctx, token)

	pr, _, err := client.PullRequests.Get(ctx, owner, name, prNumber)
	if err != nil {
		return fmt.Errorf("get pull request: %w", err)
	}

	headSHA := pr.GetHead().GetSHA()
	if headSHA == "" {
		return errors.New("pull request head SHA is empty")
	}

	commits, err := listAllCommits(ctx, client, owner, name, prNumber)
	if err != nil {
		return fmt.Errorf("list commits: %w", err)
	}

	files, err := listAllFiles(ctx, client, owner, name, prNumber)
	if err != nil {
		return fmt.Errorf("list files: %w", err)
	}

	combinedStatus, _, err := client.Repositories.GetCombinedStatus(ctx, owner, name, headSHA, nil)
	if err != nil {
		return fmt.Errorf("get combined status: %w", err)
	}

	checkRuns, err := listAllCheckRuns(ctx, client, owner, name, headSHA)
	if err != nil {
		return fmt.Errorf("list check runs: %w", err)
	}

	reviews, err := listAllReviews(ctx, client, owner, name, prNumber)
	if err != nil {
		return fmt.Errorf("list reviews: %w", err)
	}

	result := Evaluate(EvaluationInput{
		PullRequest:       pr,
		Commits:           commits,
		Files:             files,
		CombinedState:     combinedStatus.GetState(),
		CheckRuns:         checkRuns,
		CurrentRunID:      strings.TrimSpace(os.Getenv("GITHUB_RUN_ID")),
		Reviews:           reviews,
		ApproverLogin:     strings.TrimSpace(os.Getenv("GITHUB_ACTOR")),
		RepositoryOwner:   owner,
		RepositoryName:    name,
		PullRequestNumber: prNumber,
	})

	logJSON("evaluation", map[string]any{
		"owner":            owner,
		"repo":             name,
		"pull_request":     prNumber,
		"head_sha":         headSHA,
		"combined_state":   combinedStatus.GetState(),
		"check_runs_count": len(checkRuns),
		"current_run_id":   strings.TrimSpace(os.Getenv("GITHUB_RUN_ID")),
		"result":           result,
	})

	if !result.ShouldApprove {
		return nil
	}

	approvalReason := "all category and shared gates passed"
	if result.CategoryMatched != "" {
		approvalReason = fmt.Sprintf("category=%s with shared checks gates passing", result.CategoryMatched)
	}
	review := &github.PullRequestReviewRequest{
		Event: github.Ptr("APPROVE"),
		Body:  github.Ptr("Auto-approved by policy: " + approvalReason + "."),
	}

	if _, _, err := client.PullRequests.CreateReview(ctx, owner, name, prNumber, review); err != nil {
		return fmt.Errorf("create approval review: %w", err)
	}

	logJSON("approved", map[string]any{
		"owner":        owner,
		"repo":         name,
		"pull_request": prNumber,
	})

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

func readPullRequestNumber(eventPath string) (int, error) {
	if eventPath == "" {
		return 0, errors.New("missing GITHUB_EVENT_PATH")
	}

	content, err := os.ReadFile(eventPath)
	if err != nil {
		return 0, fmt.Errorf("read event payload: %w", err)
	}

	var payload eventPayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return 0, fmt.Errorf("unmarshal event payload: %w", err)
	}

	if payload.PullRequest != nil && payload.PullRequest.Number > 0 {
		return payload.PullRequest.Number, nil
	}

	if payload.CheckSuite != nil {
		for _, pr := range payload.CheckSuite.PullRequests {
			if pr.Number > 0 {
				return pr.Number, nil
			}
		}
	}

	return 0, nil
}

func listAllCommits(ctx context.Context, client *github.Client, owner, repo string, number int) ([]*github.RepositoryCommit, error) {
	opts := &github.ListOptions{PerPage: 100}
	commits := make([]*github.RepositoryCommit, 0)

	for {
		pageCommits, resp, err := client.PullRequests.ListCommits(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		commits = append(commits, pageCommits...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return commits, nil
}

func listAllFiles(ctx context.Context, client *github.Client, owner, repo string, number int) ([]*github.CommitFile, error) {
	opts := &github.ListOptions{PerPage: 100}
	files := make([]*github.CommitFile, 0)

	for {
		pageFiles, resp, err := client.PullRequests.ListFiles(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		files = append(files, pageFiles...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return files, nil
}

func listAllCheckRuns(ctx context.Context, client *github.Client, owner, repo, sha string) ([]*github.CheckRun, error) {
	opts := &github.ListCheckRunsOptions{
		Filter:      github.Ptr("latest"),
		ListOptions: github.ListOptions{PerPage: 100},
	}
	runs := make([]*github.CheckRun, 0)

	for {
		pageRuns, resp, err := client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, opts)
		if err != nil {
			return nil, err
		}
		runs = append(runs, pageRuns.CheckRuns...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return runs, nil
}

func listAllReviews(ctx context.Context, client *github.Client, owner, repo string, number int) ([]*github.PullRequestReview, error) {
	opts := &github.ListOptions{PerPage: 100}
	reviews := make([]*github.PullRequestReview, 0)

	for {
		pageReviews, resp, err := client.PullRequests.ListReviews(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, pageReviews...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return reviews, nil
}

func logJSON(kind string, payload map[string]any) {
	payload["event"] = kind
	encoded, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf(`{"event":"log_encode_error","error":%q}`+"\n", err.Error())
		return
	}
	fmt.Println(string(encoded))
}
