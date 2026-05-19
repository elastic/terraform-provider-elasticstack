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

package prmgmt_test

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/prmgmt"
	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/semver"
)

func TestResolveReleaseMode_prepReleaseBranch(t *testing.T) {
	t.Parallel()
	got := prmgmt.ResolveReleaseMode(prmgmt.EventPullRequest, "prep-release-1.2.3")
	want := prmgmt.ReleaseModeResolution{
		Mode:          prmgmt.WorkflowModeRelease,
		TargetVersion: "1.2.3",
		TargetBranch:  "prep-release-1.2.3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected resolution: %+v want %+v", got, want)
	}
}

func TestResolveReleaseMode_defaultsUnreleased(t *testing.T) {
	t.Parallel()
	got := prmgmt.ResolveReleaseMode("workflow_dispatch", testPullRequestMainBase)
	want := prmgmt.ReleaseModeResolution{
		Mode:          prmgmt.WorkflowModeUnreleased,
		TargetVersion: "",
		TargetBranch:  "generated-changelog",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected resolution: %+v want %+v", got, want)
	}
}

func TestParseSemverTagsFromRaw_filtersStrictSemverTags(t *testing.T) {
	t.Parallel()
	got := semver.ParseSemverTagsFromRaw("v1.2.3\nv1.2.3-rc1\nfoo\nv2.0.0\n")
	want := []semver.Tag{"v1.2.3", "v2.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestBuildCompareRange(t *testing.T) {
	t.Parallel()
	if got := semver.BuildCompareRange(""); got != "HEAD" {
		t.Fatalf("empty: got %q", got)
	}
	if got := semver.BuildCompareRange("v1.2.2"); got != "v1.2.2..HEAD" {
		t.Fatalf("tag: got %q", got)
	}
}

func TestSelectPreviousTag_excludesReleaseTagWhenPresent(t *testing.T) {
	t.Parallel()
	tags := []semver.Tag{"v1.2.3", "v1.2.2", "v1.2.1"}
	got := semver.SelectPreviousTag(tags, prmgmt.WorkflowModeRelease, "1.2.3")
	want := semver.PreviousTagResult{
		PreviousTag:        "v1.2.2",
		ExcludedTag:        "v1.2.3",
		ExcludedCurrentTag: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestBuildReleaseContext_combinesModeAndCompareRange(t *testing.T) {
	t.Parallel()
	got := prmgmt.BuildReleaseContext(prmgmt.EventPullRequestTarget, "prep-release-2.0.0", []semver.Tag{"v2.0.0", "v1.9.0"})
	if got.Mode != prmgmt.WorkflowModeRelease ||
		got.TargetVersion != "2.0.0" ||
		got.TargetBranch != "prep-release-2.0.0" {
		t.Fatalf("mode fields: %+v", got.ReleaseModeResolution)
	}
	if got.PreviousTag != "v1.9.0" || got.ExcludedTag != "v2.0.0" || !got.ExcludedCurrentTag {
		t.Fatalf("previous tag fields: %+v", got.PreviousTagResult)
	}
	if got.CompareRange != "v1.9.0..HEAD" {
		t.Fatalf("compare range: got %q", got.CompareRange)
	}
}
