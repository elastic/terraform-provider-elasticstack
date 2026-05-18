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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ReadEventPayload reads the raw JSON payload from path (typically
// $GITHUB_EVENT_PATH).
func ReadEventPayload(path string) ([]byte, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("missing GITHUB_EVENT_PATH")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read event payload: %w", err)
	}
	return content, nil
}

// LoadEvent decodes the JSON workflow event at path into a typed value T.
func LoadEvent[T any](path string) (T, error) {
	var zero T
	raw, err := ReadEventPayload(path)
	if err != nil {
		return zero, err
	}
	if err := json.Unmarshal(raw, &zero); err != nil {
		return zero, fmt.Errorf("unmarshal event payload: %w", err)
	}
	return zero, nil
}

// DecodeEvent unmarshals raw JSON workflow event bytes into T.
func DecodeEvent[T any](data []byte) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("unmarshal event payload: %w", err)
	}
	return v, nil
}

type pullRequestEventPayload struct {
	PullRequest *struct {
		Number int `json:"number"`
	} `json:"pull_request"`
}

// OptionalPullRequestNumberFromEventPath returns payload.pull_request.number when the
// field is present. An empty path returns (0, nil). This mirrors optional chaining on
// context.payload.pull_request?.number used by changelog workflow scripts.
func OptionalPullRequestNumberFromEventPath(path string) (int, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read event payload: %w", err)
	}
	var p pullRequestEventPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return 0, fmt.Errorf("unmarshal event payload: %w", err)
	}
	if p.PullRequest == nil || p.PullRequest.Number <= 0 {
		return 0, nil
	}
	return p.PullRequest.Number, nil
}
