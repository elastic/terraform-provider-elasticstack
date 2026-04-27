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

// Package kibanautil provides shared helpers for Kibana API clients.
package kibanautil

import (
	"context"
	"fmt"
	"net/http"
)

// BuildSpaceAwarePath constructs an API path with space awareness.
// If spaceID is empty or "default", returns the basePath unchanged.
// Otherwise, prepends "/s/{spaceID}" to the basePath.
func BuildSpaceAwarePath(spaceID, basePath string) string {
	if spaceID != "" && spaceID != "default" {
		return fmt.Sprintf("/s/%s%s", spaceID, basePath)
	}
	return basePath
}

// SpaceAwarePathRequestEditor returns a RequestEditorFn that modifies the
// request path for space awareness.
func SpaceAwarePathRequestEditor(spaceID string) func(ctx context.Context, req *http.Request) error {
	return func(_ context.Context, req *http.Request) error {
		req.URL.Path = BuildSpaceAwarePath(spaceID, req.URL.Path)
		return nil
	}
}
