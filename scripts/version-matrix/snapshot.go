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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type snapshotLatest struct {
	Version string `json:"version"`
}

func FetchSnapshotLabel(ctx context.Context, client *http.Client, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("snapshot request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("snapshot fetch: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("snapshot read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("snapshot endpoint returned %s", resp.Status)
	}
	return ParseSnapshotLabel(body)
}

func ParseSnapshotLabel(body []byte) (string, error) {
	var payload snapshotLatest
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("parse snapshot label: %w", err)
	}
	if strings.TrimSpace(payload.Version) == "" {
		return "", fmt.Errorf("snapshot response missing version")
	}
	return payload.Version, nil
}
