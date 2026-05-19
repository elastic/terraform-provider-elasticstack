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

package githubx_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
)

func TestNewGitHubClient_tokenFromEnv(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		envValue    string
		wantErr     bool
		errContains string
		wantClient  bool
	}{
		{
			name:        "empty env",
			envValue:    "",
			wantErr:     true,
			errContains: "GITHUB_TOKEN",
		},
		{
			name:        "whitespace only",
			envValue:    "  \t\n  ",
			wantErr:     true,
			errContains: "GITHUB_TOKEN",
		},
		{
			name:       "non-empty token",
			envValue:   "test-token",
			wantClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", tt.envValue)
			token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
			client, err := githubx.NewGitHubClient(ctx, token)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q should mention %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !tt.wantClient || client == nil {
				t.Fatal("expected non-nil client")
			}
		})
	}
}
