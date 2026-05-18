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

// Package prcheck validates pull request bodies against the repository's
// "## Changelog" contract (parity with the migrated JavaScript verifier workflow).
//
// Packages in prcheck MUST NOT read environment variables or touch the filesystem;
// callers inject GitHub/network via PullRequestFetcher.
package prcheck

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
)

const (
	// StatusPass is written to Verdict.Status when validation succeeds or no-changelog bypass applies.
	StatusPass = "pass"
	// StatusFail is written to Verdict.Status when validation reports one or more errors.
	StatusFail = "fail"

	defaultNoChangelogLabel = "no-changelog"
)

// Verdict is the canonical JSON verdict for workflows (see GitHub Actions result_json consumer).
type Verdict struct {
	Status Status `json:"status"`

	// Errors lists validation diagnostics when Status is StatusFail (JS errors array parity).
	Errors []string `json:"errors,omitempty"`

	// NoChangelogSkip is true when validation was bypassed via the configured no-changelog label.
	NoChangelogSkip bool `json:"no_changelog_skip,omitempty"`
}

// Status is serialized as the JSON status string ("pass" | "fail").
type Status string

func (s Status) String() string { return string(s) }

// MarshalJSON encodes Status as a JSON string literal.
func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(s) + `"`), nil
}

// UnmarshalJSON decodes Status from JSON string literals ("pass", "fail").
func (s *Status) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("prcheck: status must be a JSON string")
	}
	st := Status(data[1 : len(data)-1])
	switch st {
	case StatusPass, StatusFail:
		*s = st
		return nil
	default:
		return fmt.Errorf("prcheck: invalid status JSON value %s", string(data))
	}
}

// PullRequest is the validator input trimmed from GitHub PR APIs.
type PullRequest struct {
	Number int
	Body   string
	Labels []string
}

// PullRequestFetcher retrieves PR payload for Validate.
type PullRequestFetcher interface {
	GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error)
}

// ValidateOptions configures Validate (no environment access).
type ValidateOptions struct {
	Owner            string
	Repo             string
	Number           int
	Fetcher          PullRequestFetcher
	NoChangelogLabel string
}

// Validate loads the PR, applies optional no-changelog suppression, parses via section.Parse,
// and validates via section.ValidateChangelogSectionFull (default options).
//
// Operational errors from Fetcher bubble up wrapped; validation failures return Verdict.StatusFail
// and a nil error.
func Validate(ctx context.Context, opts ValidateOptions) (Verdict, error) {
	switch {
	case opts.Fetcher == nil:
		return Verdict{}, errors.New("fetcher required")
	case opts.Owner == "" || opts.Repo == "":
		return Verdict{}, errors.New("owner and repo are required")
	case opts.Number <= 0:
		return Verdict{}, errors.New("pull request number must be positive")
	}

	label := opts.NoChangelogLabel
	if label == "" {
		label = defaultNoChangelogLabel
	}

	pr, err := opts.Fetcher.GetPullRequest(ctx, opts.Owner, opts.Repo, opts.Number)
	if err != nil {
		return Verdict{}, fmt.Errorf("fetch pull request #%d: %w", opts.Number, err)
	}

	if slices.Contains(pr.Labels, label) {
		return Verdict{
			Status:          StatusPass,
			NoChangelogSkip: true,
		}, nil
	}

	sec, parseErr := section.Parse([]byte(pr.Body))
	var parsedPtr *section.Section
	switch {
	case parseErr != nil && errors.Is(parseErr, section.ErrNoChangelogSection):
		parsedPtr = nil
	case parseErr != nil:
		return Verdict{}, fmt.Errorf("parse changelog section: %w", parseErr)
	default:
		secCopy := sec
		parsedPtr = &secCopy
	}

	valid, errs := section.ValidateChangelogSectionFull(parsedPtr, section.ValidateOpts{})
	if valid {
		return Verdict{Status: StatusPass}, nil
	}
	return Verdict{Status: StatusFail, Errors: errs}, nil
}
