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

package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPutSnapshotRepository_S3RawBody(t *testing.T) {
	t.Parallel()

	settings := map[string]any{
		"bucket":            "test-bucket",
		"client":            "default",
		"compress":          true,
		"readonly":          false,
		"endpoint":          "https://minio.example.com:9000",
		"path_style_access": true,
	}

	var capturedBody string
	var captureErr error
	srv := newMockElasticsearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/_snapshot/my-repo") {
			var body []byte
			body, captureErr = io.ReadAll(r.Body)
			capturedBody = string(body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"acknowledged":true}`)
	})
	defer srv.Close()

	client := newMockScopedClient(t, srv)
	diags := PutSnapshotRepository(context.Background(), client, "my-repo", "s3", settings, false)
	require.NoError(t, captureErr)
	require.False(t, diags.HasError(), diags.Errors())
	require.NotEmpty(t, capturedBody)

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(capturedBody), &body))
	require.Equal(t, "s3", body["type"])

	settingsBody, ok := body["settings"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "https://minio.example.com:9000", settingsBody["endpoint"])
	require.Equal(t, true, settingsBody["path_style_access"])
	require.Equal(t, "test-bucket", settingsBody["bucket"])
}

func TestPutSnapshotRepository_S3RawBodyWithoutEndpoint(t *testing.T) {
	t.Parallel()

	settings := map[string]any{
		"bucket":            "test-bucket",
		"client":            "default",
		"compress":          true,
		"readonly":          false,
		"path_style_access": false,
	}

	var capturedBody string
	var captureErr error
	srv := newMockElasticsearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/_snapshot/my-repo") {
			var body []byte
			body, captureErr = io.ReadAll(r.Body)
			capturedBody = string(body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"acknowledged":true}`)
	})
	defer srv.Close()

	client := newMockScopedClient(t, srv)
	diags := PutSnapshotRepository(context.Background(), client, "my-repo", "s3", settings, false)
	require.NoError(t, captureErr)
	require.False(t, diags.HasError(), diags.Errors())
	require.NotEmpty(t, capturedBody)

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(capturedBody), &body))

	settingsBody, ok := body["settings"].(map[string]any)
	require.True(t, ok)
	require.NotContains(t, settingsBody, "endpoint")
}

func TestPutSnapshotRepository_HDFSRawBody(t *testing.T) {
	t.Parallel()

	settings := map[string]any{
		"uri":           "hdfs://namenode:8020",
		"path":          "/repos/snapshots",
		"load_defaults": true,
		"compress":      true,
		"readonly":      false,
	}

	var capturedBody string
	var captureErr error
	srv := newMockElasticsearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/_snapshot/hdfs-repo") {
			var body []byte
			body, captureErr = io.ReadAll(r.Body)
			capturedBody = string(body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"acknowledged":true}`)
	})
	defer srv.Close()

	client := newMockScopedClient(t, srv)
	diags := PutSnapshotRepository(context.Background(), client, "hdfs-repo", "hdfs", settings, false)
	require.NoError(t, captureErr)
	require.False(t, diags.HasError(), diags.Errors())
	require.NotEmpty(t, capturedBody)

	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(capturedBody), &body))
	require.Equal(t, "hdfs", body["type"])

	settingsBody, ok := body["settings"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "hdfs://namenode:8020", settingsBody["uri"])
	require.Equal(t, "/repos/snapshots", settingsBody["path"])
	require.Equal(t, true, settingsBody["load_defaults"])
}
