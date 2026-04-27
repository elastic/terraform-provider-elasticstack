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

package kibanautil_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
)

func TestBuildSpaceAwarePath(t *testing.T) {
	tests := []struct {
		name     string
		spaceID  string
		basePath string
		want     string
	}{
		{
			name:     "empty spaceID returns basePath unchanged",
			spaceID:  "",
			basePath: "/api/fleet/outputs",
			want:     "/api/fleet/outputs",
		},
		{
			name:     "default spaceID returns basePath unchanged",
			spaceID:  "default",
			basePath: "/api/fleet/outputs",
			want:     "/api/fleet/outputs",
		},
		{
			name:     "custom spaceID prepends /s/{spaceID}",
			spaceID:  "my-space",
			basePath: "/api/fleet/outputs",
			want:     "/s/my-space/api/fleet/outputs",
		},
		{
			name:     "custom spaceID with empty basePath",
			spaceID:  "my-space",
			basePath: "",
			want:     "/s/my-space",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := kibanautil.BuildSpaceAwarePath(tc.spaceID, tc.basePath)
			if got != tc.want {
				t.Errorf("BuildSpaceAwarePath(%q, %q) = %q, want %q", tc.spaceID, tc.basePath, got, tc.want)
			}
		})
	}
}

func TestSpaceAwarePathRequestEditor(t *testing.T) {
	tests := []struct {
		name     string
		spaceID  string
		initPath string
		wantPath string
	}{
		{
			name:     "empty spaceID leaves path unchanged",
			spaceID:  "",
			initPath: "/api/dashboards/dashboard/abc",
			wantPath: "/api/dashboards/dashboard/abc",
		},
		{
			name:     "default spaceID leaves path unchanged",
			spaceID:  "default",
			initPath: "/api/dashboards/dashboard/abc",
			wantPath: "/api/dashboards/dashboard/abc",
		},
		{
			name:     "custom spaceID prefixes path",
			spaceID:  "ops",
			initPath: "/api/dashboards/dashboard/abc",
			wantPath: "/s/ops/api/dashboards/dashboard/abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			editor := kibanautil.SpaceAwarePathRequestEditor(tc.spaceID)
			req := &http.Request{URL: &url.URL{Path: tc.initPath}}
			if err := editor(context.Background(), req); err != nil {
				t.Fatalf("editor returned unexpected error: %v", err)
			}
			if req.URL.Path != tc.wantPath {
				t.Errorf("after editor, path = %q, want %q", req.URL.Path, tc.wantPath)
			}
		})
	}
}
