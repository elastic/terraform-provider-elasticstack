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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSnapshotLabelFromMasterJSON(t *testing.T) {
	t.Parallel()

	body := []byte(`{
 "version" : "9.6.0-SNAPSHOT",
 "build_id" : "9.6.0-8f8a7ebc",
 "manifest_url" : "https://snapshots.elastic.co/9.6.0-8f8a7ebc/manifest-9.6.0-SNAPSHOT.json"
}`)
	got, err := ParseSnapshotLabel(body)
	require.NoError(t, err)
	assert.Equal(t, "9.6.0-SNAPSHOT", got)
}

func TestParseSnapshotLabelRejectsNonSnapshotVersion(t *testing.T) {
	t.Parallel()

	_, err := ParseSnapshotLabel([]byte(`{"version":"9.6.0"}`))
	require.Error(t, err)
}

func TestParseSnapshotLabelRejectsUnparsableJSON(t *testing.T) {
	t.Parallel()

	_, err := ParseSnapshotLabel([]byte(`{"version":`))
	require.Error(t, err)
}

func TestFetchSnapshotLabelFailsWhenEndpointUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	_, err := FetchSnapshotLabel(context.Background(), server.Client(), server.URL)
	require.Error(t, err)
}

func TestFetchSnapshotLabelFailsWhenVersionMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"build_id":"9.6.0-8f8a7ebc"}`))
	}))
	t.Cleanup(server.Close)

	_, err := FetchSnapshotLabel(context.Background(), server.Client(), server.URL)
	require.Error(t, err)
}
