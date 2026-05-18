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

package semver_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/semver"
)

func TestBuildCompareRange_emptyPrevious(t *testing.T) {
	if g := semver.BuildCompareRange(""); g != "HEAD" {
		t.Fatalf("got %q", g)
	}
}

func TestBuildCompareRange_withPrevious(t *testing.T) {
	if g := semver.BuildCompareRange("v1.9.0"); g != "v1.9.0..HEAD" {
		t.Fatalf("got %q", g)
	}
}

func TestSelectPreviousTag_releaseExcludesCurrentTag(t *testing.T) {
	tags := []semver.Tag{"v2.0.0", "v1.9.0"}
	sel := semver.SelectPreviousTag(tags, "release", "2.0.0")
	if sel.PreviousTag != "v1.9.0" || !sel.ExcludedCurrentTag || sel.ExcludedTag != "v2.0.0" {
		t.Fatalf("got %+v", sel)
	}
}

func TestSelectPreviousTag_unreleasedKeepsLatest(t *testing.T) {
	tags := []semver.Tag{"v2.0.0", "v1.9.0"}
	sel := semver.SelectPreviousTag(tags, "unreleased", "")
	if sel.PreviousTag != "v2.0.0" || sel.ExcludedTag != "" || sel.ExcludedCurrentTag {
		t.Fatalf("got %+v", sel)
	}
}

func TestResolveCompareContext_unreleasedNoTags(t *testing.T) {
	t.Parallel()
	tags := semver.SelectPreviousTag(nil, "unreleased", "")
	if tags.PreviousTag != "" {
		t.Fatalf("want empty previous tag")
	}
	if g := semver.BuildCompareRange(tags.PreviousTag); g != "HEAD" {
		t.Fatalf("compare %q", g)
	}
}

func TestResolveCompareContext_releaseExcludesVersion(t *testing.T) {
	t.Parallel()
	raw := []semver.Tag{"v2.0.0", "v1.9.0"}
	sel := semver.SelectPreviousTag(raw, "release", "2.0.0")
	if sel.PreviousTag != "v1.9.0" {
		t.Fatalf("previous %q", sel.PreviousTag)
	}
	rng := semver.BuildCompareRange(sel.PreviousTag)
	if rng != "v1.9.0..HEAD" {
		t.Fatalf("compare %q", rng)
	}
}
