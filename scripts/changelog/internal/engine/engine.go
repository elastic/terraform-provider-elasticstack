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

package engine

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/rewriter"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/semver"
)

const prepReleaseBranchFmt = "prep-release-%s"
const changelogFilePerm = 0o644

// FS abstracts changelog file IO for Run.
type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
}

// MergedPRGatherer lists merged Pull Requests reachable from the git compare range (mirrors JS engine).
// The second return value is soft-failure warnings (mirrors core.warning in the JS engine).
type MergedPRGatherer interface {
	GatherMergedPRs(ctx context.Context, owner, repo, compareRange string) ([]section.MergedPR, []string, error)
}

// Options configures engine.Run — callers supply env-derived fields and adapters.
type Options struct {
	Mode          string
	TargetVersion string
	Owner         string
	Repo          string
	ChangelogPath string
	Now           func() time.Time
	FS            FS
	Git           semver.Execer
	Gather        MergedPRGatherer
}

// Result summarizes engine execution for callers and CI outputs.
type Result struct {
	TargetVersionOutput    string // empty when not release (matches JS target_version output)
	PreviousTag            string
	CompareRange           string
	TargetBranch           string // resolved canonical branch prior to TARGET_BRANCH override
	SectionHeader          string
	HasPRs                 bool
	HasUserFacingChanges   bool
	Included               []section.IncludedPR
	Excluded               []section.ExcludedPR
	SkippedChangelogUpdate bool
	Warnings               []string
}

// Run validates inputs, resolves compare context, gathers PRs, renders, and optionally rewrites CHANGELOG.md.
func Run(ctx context.Context, opts Options) (Result, error) {
	if err := ValidateModeAndTargetVersion(opts.Mode, opts.TargetVersion); err != nil {
		return Result{}, err
	}
	if opts.Gather == nil {
		return Result{}, errors.New("engine: Gather must be non-nil")
	}
	if opts.FS == nil {
		return Result{}, errors.New("engine: FS must be non-nil")
	}
	if opts.Now == nil {
		return Result{}, errors.New("engine: Now must be non-nil")
	}

	var tags []semver.Tag
	warnMsgs := []string{}
	if opts.Git != nil {
		var terr error
		tags, terr = semver.ListReleaseTags(opts.Git)
		if terr != nil {
			tags = nil
			warnMsgs = append(warnMsgs, fmt.Sprintf("Failed to list git tags: %v", terr))
		}
	}

	tagPick := semver.SelectPreviousTag(tags, opts.Mode, opts.TargetVersion)
	compareRange := semver.BuildCompareRange(tagPick.PreviousTag)

	var targetBranch string
	switch opts.Mode {
	case ModeUnreleased:
		targetBranch = branchGeneratedChangelog
	case ModeRelease:
		targetBranch = fmt.Sprintf(prepReleaseBranchFmt, opts.TargetVersion)
	default:
		return Result{}, fmt.Errorf("engine: unknown mode %q", opts.Mode)
	}

	records, gatherWarns, gerr := opts.Gather.GatherMergedPRs(ctx, opts.Owner, opts.Repo, compareRange)
	if gerr != nil {
		return Result{}, fmt.Errorf("gather merged pull requests: %w", gerr)
	}
	warnMsgs = append(warnMsgs, gatherWarns...)

	hasPRs := len(records) > 0
	renderChangelog := opts.Mode == ModeRelease || hasPRs

	tvOut := ""
	if opts.Mode == ModeRelease {
		tvOut = opts.TargetVersion
	}

	dupWarn := func() []string { return append([]string(nil), warnMsgs...) }

	res := Result{
		TargetVersionOutput:  tvOut,
		PreviousTag:          tagPick.PreviousTag,
		CompareRange:         compareRange,
		TargetBranch:         targetBranch,
		SectionHeader:        headerUnreleasedMarkdown,
		HasPRs:               hasPRs,
		HasUserFacingChanges: false,
		Warnings:             dupWarn(),
	}

	if !renderChangelog {
		res.SkippedChangelogUpdate = true
		return res, nil
	}

	current, readErr := opts.FS.ReadFile(opts.ChangelogPath)
	currentStr := ""
	switch {
	case readErr == nil:
		currentStr = string(current)
	case errors.Is(readErr, fs.ErrNotExist):
		warnMsgs = append(warnMsgs, fmt.Sprintf(
			"Could not read %s: %v. Will create a new file.",
			opts.ChangelogPath,
			readErr,
		))
		res.Warnings = dupWarn()
	default:
		return Result{}, fmt.Errorf("read changelog %s: %w", opts.ChangelogPath, readErr)
	}

	renderOutcome := section.RenderChangelogSection(records)
	if !renderOutcome.Success {
		return Result{}, fmt.Errorf("%s", FormatAssemblyFailureMessage(renderOutcome.Errors))
	}

	dateStr := opts.Now().UTC().Format("2006-01-02")

	var rewriteMode rewriter.RewriteMode
	var rewriteHeader string
	var rv string
	switch opts.Mode {
	case ModeRelease:
		rewriteMode = rewriter.ModeRelease
		rv = opts.TargetVersion
		res.SectionHeader = fmt.Sprintf("## [%s] - %s", opts.TargetVersion, dateStr)
		rewriteHeader = fmt.Sprintf("[%s] - %s", opts.TargetVersion, dateStr)
	default:
		rewriteMode = rewriter.ModeUnreleased
		res.SectionHeader = headerUnreleasedMarkdown
		rewriteHeader = "[Unreleased]"
	}

	body := renderOutcome.SectionBody
	if body != "" {
		body = "\n" + body
	}

	rewrite := rewriter.SectionRewrite{
		Header: rewriteHeader,
		Body:   body,
	}

	updated, werr := rewriter.RewriteSection([]byte(currentStr), rewrite, rewriteMode, rv)
	if werr != nil {
		return Result{}, fmt.Errorf("rewrite changelog section: %w", werr)
	}

	if opts.Mode == ModeRelease && opts.TargetVersion != "" && tagPick.PreviousTag != "" {
		var lnErr error
		updated, lnErr = rewriter.UpdateLinks(updated, []rewriter.LinkEntry{
			{TargetVersion: opts.TargetVersion, PreviousTag: tagPick.PreviousTag},
		})
		if lnErr != nil {
			return Result{}, fmt.Errorf("update changelog compare links: %w", lnErr)
		}
	}

	if err := opts.FS.WriteFile(opts.ChangelogPath, updated, changelogFilePerm); err != nil {
		return Result{}, fmt.Errorf("failed to write %s: %w", opts.ChangelogPath, err)
	}

	res.Included = renderOutcome.Included
	res.Excluded = renderOutcome.Excluded
	res.HasUserFacingChanges = strings.TrimSpace(renderOutcome.SectionBody) != ""
	res.Warnings = dupWarn()

	return res, nil
}
