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

package engine_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/engine"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
)

const (
	fixtureOwnerSlug = "o"
	fixtureRepoSlug  = "r"
)

type fixedGather struct {
	recs []section.MergedPR
}

func (f fixedGather) GatherMergedPRs(context.Context, string, string, string) ([]section.MergedPR, error) {
	return f.recs, nil
}

type tagOnlyGit struct {
	tags string
}

func (g tagOnlyGit) Run(name string, args ...string) ([]byte, error) {
	_, _ = name, args
	if len(args) >= 1 && args[0] == "tag" {
		return []byte(g.tags), nil
	}
	return nil, nil
}

func fixedNow(tb testing.TB, m time.Month, d int) func() time.Time {
	tb.Helper()
	const y = 2026
	return func() time.Time {
		return time.Date(y, m, d, 12, 0, 0, 0, time.UTC)
	}
}

func mustRun(ctx context.Context, tb testing.TB, opts engine.Options) engine.Result {
	tb.Helper()
	res, err := engine.Run(ctx, opts)
	if err != nil {
		tb.Fatal(err)
	}
	return res
}

func TestFormatAssemblyFailureMessage_listsReasons(t *testing.T) {
	msg := engine.FormatAssemblyFailureMessage([]section.AssemblyError{{Reason: "bad PR"}})
	if !strings.Contains(msg, "bad PR") || !strings.Contains(msg, "Changelog assembly failed") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestRunRenderAndWrite_setsHasUserFacingChanges(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	body := strings.Join([]string{
		"# L", "", "## [Unreleased]", "old", "", "## [0.1.0]", "x",
	}, "\n")
	if err := os.WriteFile(changelogPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeUnreleased,
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 5, 18),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{},
		Gather: fixedGather{recs: []section.MergedPR{{
			Number: 1,
			URL:    "https://github.com/o/r/pull/1",
			Labels: nil,
			Body:   "## Changelog\nCustomer impact: fix\nSummary: hello\n",
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasPRs || !res.HasUserFacingChanges {
		t.Fatalf("unexpected flags: %+v", res)
	}
	text, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "hello") || strings.Contains(string(text), "\nold\n") {
		t.Fatalf("unexpected changelog:\n%s", text)
	}
}

func TestRunRenderAndWrite_breakingOnlyImpact(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(changelogPath, []byte("# L\n\n## [Unreleased]\nold\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prBody := strings.Join([]string{
		"## Changelog",
		"Customer impact: breaking",
		"Summary: A breaking change",
		"",
		"### Breaking changes",
		"A new required env var `FOO` must be set.",
	}, "\n")
	res, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeRelease,
		TargetVersion: "1.0.0",
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 3, 10),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{},
		Gather: fixedGather{recs: []section.MergedPR{{
			Number: 7,
			URL:    "https://github.com/o/r/pull/7",
			Labels: nil,
			Body:   prBody,
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Included) != 1 || !res.HasPRs || !res.HasUserFacingChanges {
		t.Fatalf("unexpected result: %+v", res)
	}
	text := string(slurp(t, changelogPath))
	if !strings.Contains(text, "### Breaking changes") || !strings.Contains(text, "FOO") {
		t.Fatalf("unexpected changelog:\n%s", text)
	}
}

func TestRunRenderAndWrite_breakingImpactNoneExcludedButRendered(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(changelogPath, []byte("# L\n\n## [Unreleased]\nold\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prBody := strings.Join([]string{
		"## Changelog",
		"Customer impact: none",
		"",
		"### Breaking changes",
		"Internal schema change with no API surface impact.",
	}, "\n")
	res, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeRelease,
		TargetVersion: "1.0.0",
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 3, 10),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{},
		Gather: fixedGather{recs: []section.MergedPR{{
			Number: 8,
			URL:    "https://github.com/o/r/pull/8",
			Labels: nil,
			Body:   prBody,
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Included) != 0 || len(res.Excluded) != 1 || !res.HasPRs || !res.HasUserFacingChanges {
		t.Fatalf("unexpected result: included=%d excluded=%d %+v",
			len(res.Included), len(res.Excluded), res)
	}
	text := string(slurp(t, changelogPath))
	if !strings.Contains(text, "### Breaking changes") || !strings.Contains(text, "Internal schema change") {
		t.Fatalf("unexpected changelog:\n%s", text)
	}
}

func TestRunRenderAndWrite_releaseReplacesUnreleased(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	body := strings.Join([]string{
		"# L", "", "## [Unreleased]", "pending", "", "## [0.9.0]", "z", "",
		"[Unreleased]: https://github.com/o/r/compare/v0.9.0...HEAD",
		"[0.9.0]: https://github.com/o/r/compare/v0.8.0...v0.9.0",
	}, "\n")
	if err := os.WriteFile(changelogPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeRelease,
		TargetVersion: "1.0.0",
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 2, 1),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{tags: "v0.9.0\n"}, // excludes nothing for 1.0.0; previous tag v0.9.0
		Gather: fixedGather{recs: []section.MergedPR{{
			Number: 2,
			URL:    "https://github.com/o/r/pull/2",
			Labels: nil,
			Body:   "## Changelog\nCustomer impact: enhancement\nSummary: ship\n",
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	headerRe := regexp.MustCompile(`^## \[1\.0\.0\] - 2026-02-01$`)
	if !headerRe.MatchString(res.SectionHeader) {
		t.Fatalf("unexpected header %q", res.SectionHeader)
	}
	text := string(slurp(t, changelogPath))
	if strings.Contains(text, "## [Unreleased]") || !strings.Contains(text, "ship") {
		t.Fatalf("unexpected changelog:\n%s", text)
	}
	wantUnrel := "[Unreleased]: https://github.com/o/r/compare/v1.0.0...HEAD"
	wantRel := "[1.0.0]: https://github.com/o/r/compare/v0.9.0...v1.0.0"
	if !strings.Contains(text, wantUnrel) || !strings.Contains(text, wantRel) {
		t.Fatalf("missing link rows; got:\n%s", text)
	}
	r := strings.Index(text, "## [1.0.0]")
	old := strings.Index(text, "## [0.9.0]")
	if r == -1 || old == -1 || r >= old {
		t.Fatalf("section order wrong: indices %d %d", r, old)
	}
}

func TestRunRenderAndWrite_releaseZeroPRsClearsLinks(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	body := strings.Join([]string{
		"# L", "", "## [Unreleased]", "pending", "", "## [0.9.0]", "z", "",
		"[Unreleased]: https://github.com/o/r/compare/v0.9.0...HEAD",
		"[0.9.0]: https://github.com/o/r/compare/v0.8.0...v0.9.0",
	}, "\n")
	if err := os.WriteFile(changelogPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeRelease,
		TargetVersion: "1.0.0",
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 7, 4),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{tags: "v0.9.0\n"},
		Gather:        fixedGather{recs: nil},
	})
	if err != nil {
		t.Fatal(err)
	}
	headerRe := regexp.MustCompile(`^## \[1\.0\.0\] - 2026-07-04$`)
	if !headerRe.MatchString(res.SectionHeader) {
		t.Fatalf("unexpected header %q", res.SectionHeader)
	}
	if res.HasPRs || res.HasUserFacingChanges || len(res.Included) != 0 {
		t.Fatalf("unexpected flags / included: %+v", res)
	}
	text := string(slurp(t, changelogPath))
	if strings.Contains(text, "## [Unreleased]") || strings.Contains(text, "pending") {
		t.Fatalf("unexpected changelog:\n%s", text)
	}
	wantUnrel := "[Unreleased]: https://github.com/o/r/compare/v1.0.0...HEAD"
	wantRel := "[1.0.0]: https://github.com/o/r/compare/v0.9.0...v1.0.0"
	if !strings.Contains(text, wantUnrel) || !strings.Contains(text, wantRel) {
		t.Fatalf("missing link rows; got:\n%s", text)
	}
}

func TestRunRenderAndWrite_assemblyFails(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(changelogPath, []byte("# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeUnreleased,
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 1, 1),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{},
		Gather: fixedGather{recs: []section.MergedPR{{
			Number: 99,
			URL:    "https://github.com/o/r/pull/99",
			Labels: nil,
			Body:   "no changelog block",
		}}},
	})
	if err == nil || !strings.Contains(err.Error(), "Changelog assembly failed") {
		t.Fatalf("expected assembly error, got %v", err)
	}
	if !strings.Contains(err.Error(), "missing a required ## Changelog") {
		t.Fatalf("unexpected wording: %v", err)
	}
}

func TestRun_skipUnreleasedWithNoPRs(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if err := os.WriteFile(changelogPath, []byte("# untouched\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := engine.Run(ctx, engine.Options{
		Mode:          engine.ModeUnreleased,
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 1, 2),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{},
		Gather:        fixedGather{recs: nil},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.SkippedChangelogUpdate || res.HasPRs || res.HasUserFacingChanges ||
		res.SectionHeader != "## [Unreleased]" {
		t.Fatalf("unexpected %+v", res)
	}
	text := string(slurp(t, changelogPath))
	if text != "# untouched\n" {
		t.Fatalf("file mutated: %q", text)
	}
}

func TestRun_engineEndToEndFromFactoryTest(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	body := strings.Join([]string{"# L", "", "## [Unreleased]", "x", "", "## [0.1.0]", "y"}, "\n")
	if err := os.WriteFile(changelogPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	res := mustRun(ctx, t, engine.Options{
		Mode:          engine.ModeUnreleased,
		Owner:         fixtureOwnerSlug,
		Repo:          fixtureRepoSlug,
		ChangelogPath: changelogPath,
		Now:           fixedNow(t, 4, 1),
		FS:            engineOSFS{},
		Git:           tagOnlyGit{tags: "v0.1.0\n"},
		Gather: fixedGather{recs: []section.MergedPR{{
			Number: 5,
			URL:    "https://github.com/o/r/pull/5",
			Labels: nil,
			Body:   "## Changelog\nCustomer impact: enhancement\nSummary: feat done\n",
		}}},
	})

	if res.TargetVersionOutput != "" || res.PreviousTag != "v0.1.0" || res.CompareRange != "v0.1.0..HEAD" ||
		res.TargetBranch != "generated-changelog" || !res.HasPRs || !res.HasUserFacingChanges {
		t.Fatalf("unexpected result: %+v", res)
	}
	text := string(slurp(t, changelogPath))
	if !strings.Contains(text, "feat done") {
		t.Fatalf("missing rendered body:\n%s", text)
	}
}

// engineOSFS bridges os file ops to engine.FS for blackbox tests.
type engineOSFS struct{}

func (engineOSFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (engineOSFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func slurp(tb testing.TB, path string) []byte {
	tb.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		tb.Fatal(err)
	}
	return b
}
