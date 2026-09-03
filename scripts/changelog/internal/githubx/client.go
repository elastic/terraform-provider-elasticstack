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

package githubx

import (
	"context"
	"errors"
	"strings"

	"github.com/google/go-github/v89/github"
)

// NewGitHubClient returns a GitHub API client authenticated with token,
// mirroring scripts/auto-approve/main.go.
func NewGitHubClient(_ context.Context, token string) (*github.Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing GITHUB_TOKEN")
	}
	return github.NewClient(github.WithAuthToken(token))
}
