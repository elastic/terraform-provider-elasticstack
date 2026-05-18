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

package rewriter

import (
	"regexp"
	"strings"
	"testing"
)

func TestUpdateLinksStandardReleaseUpdatesUnreleasedLineAndInsertsRelease(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...HEAD",
		"[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5",
	}, "\n")

	out := mustUpdateLinks(t, before, LinkEntry{
		TargetVersion: "0.15.0",
		PreviousTag:   "v0.14.5",
	})

	outStable := mustUpdateLinks(t, string(out), LinkEntry{
		TargetVersion: "0.15.0",
		PreviousTag:   "v0.14.5",
	})
	if string(out) != string(outStable) {
		t.Fatalf("UpdateLinks should be idempotent immediately after inserting the release compare line")
	}
	s := string(out)

	okUnreleasedLine, err := regexp.MatchString(`(?m)^\[Unreleased\]: https:\/\/github\.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.15\.0\.\.\.HEAD$`, s)
	if err != nil || !okUnreleasedLine {
		t.Fatalf("want updated [Unreleased] line")
	}
	okReleaseLine, err := regexp.MatchString(`(?m)^\[0\.15\.0\]: https:\/\/github\.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.14\.5\.\.\.v0\.15\.0$`, s)
	if err != nil || !okReleaseLine {
		t.Fatalf("want inserted release link line")
	}

	orderedTail := regexp.MustCompile(`\[Unreleased\]: https://github\.com/elastic/terraform-provider-elasticstack/compare/v0\.15\.0\.\.\.HEAD\n` +
		`\[0\.15\.0\]: https://github\.com/elastic/terraform-provider-elasticstack/compare/v0\.14\.5\.\.\.v0\.15\.0\n` +
		`\[0\.14\.5\]:`)
	if !orderedTail.MatchString(s) {
		t.Fatalf("expected contiguous ordered link trio")
	}
}

func TestUpdateLinksIsIdempotentWhenReleaseEntryAlreadyExists(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.15.0...HEAD",
		"[0.15.0]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...v0.15.0",
		"[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5",
	}, "\n")

	entry := LinkEntry{TargetVersion: "0.15.0", PreviousTag: "v0.14.5"}
	first := mustUpdateLinks(t, before, entry)
	second := mustUpdateLinks(t, string(first), entry)
	if string(first) != string(second) {
		t.Fatalf("expected idempotent rewrite; first != second")
	}

	releaseCount := regexp.MustCompile(`(?m)^\[0\.15\.0\]:`).FindAllIndex(second, -1)
	if len(releaseCount) != 1 {
		t.Fatalf("want exactly one [0.15.0] link line")
	}

	okUnreleased, err := regexp.MatchString(`(?m)^\[Unreleased\]: https:\/\/github\.com\/elastic\/terraform-provider-elasticstack\/compare\/v0\.15\.0\.\.\.HEAD$`, string(second))
	if err != nil || !okUnreleased {
		t.Fatalf("[Unreleased] line should remain stable across idempotent reruns")
	}
}

func mustUpdateLinks(tb testing.TB, before string, entry LinkEntry) []byte {
	tb.Helper()
	out, err := UpdateLinks([]byte(before), []LinkEntry{entry})
	if err != nil {
		tb.Fatalf("UpdateLinks error: %v", err)
	}
	return out
}

func TestUpdateLinksIsNoOpWhenUnreleasedAbsent(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5",
	}, "\n")

	got := mustUpdateLinks(t, before, LinkEntry{TargetVersion: "0.15.0", PreviousTag: "v0.14.5"})
	if string(got) != before {
		t.Fatalf("want unchanged content")
	}
}

func TestUpdateLinksIsNoOpWhenPreviousTagMissing(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...HEAD",
		"[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5",
	}, "\n")

	got := mustUpdateLinks(t, before, LinkEntry{TargetVersion: "0.15.0"})
	if string(got) != before {
		t.Fatalf("want unchanged content")
	}
}

func TestUpdateLinksIsNoOpWhenTargetVersionMissing(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{
		"# Changelog",
		"",
		"[Unreleased]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.5...HEAD",
		"[0.14.5]: https://github.com/elastic/terraform-provider-elasticstack/compare/v0.14.4...v0.14.5",
	}, "\n")

	got := mustUpdateLinks(t, before, LinkEntry{PreviousTag: "v0.14.5"})
	if string(got) != before {
		t.Fatalf("want unchanged content")
	}
}
