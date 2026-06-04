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

package entity

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildID(t *testing.T) {
	tests := []struct {
		name     string
		spaceID  string
		entityID string
		want     string
	}{
		{
			name:     "default space",
			spaceID:  "",
			entityID: "host:web-01",
			want:     "default/host:web-01",
		},
		{
			name:     "custom space",
			spaceID:  "production",
			entityID: "host:web-01",
			want:     "production/host:web-01",
		},
		{
			name:     "entity ID with colons",
			spaceID:  "default",
			entityID: "user:john:doe",
			want:     "default/user:john:doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildID(tt.spaceID, tt.entityID)
			if got != tt.want {
				t.Errorf("buildID(%q, %q) = %q, want %q", tt.spaceID, tt.entityID, got, tt.want)
			}
		})
	}
}

func TestNormalizeSpaceID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		isNull bool
		want   string
	}{
		{
			name:  "default space from empty",
			input: "",
			want:  "default",
		},
		{
			name:  "custom space preserved",
			input: "production",
			want:  "production",
		},
		{
			name:   "null returns default",
			isNull: true,
			want:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v = types.StringValue(tt.input)
			if tt.isNull {
				v = types.StringNull()
			}
			got := NormalizeSpaceID(v)
			if got != tt.want {
				t.Errorf("NormalizeSpaceID(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
