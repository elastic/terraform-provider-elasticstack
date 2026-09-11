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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v89/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatestPatchPerMinorKeepsHighestPatch(t *testing.T) {
	t.Parallel()

	tags := []string{"v8.19.17", "v8.19.21", "v9.5.0", "v9.5.3"}
	got := LatestPatchPerMinor(tags)
	assert.Equal(t, []string{"8.19.21", "9.5.3"}, got)
}

func TestLatestPatchPerMinorExcludesMajorsOutside8And9(t *testing.T) {
	t.Parallel()

	tags := []string{"v7.17.28", "v8.19.21", "v10.0.0"}
	got := LatestPatchPerMinor(tags)
	assert.Equal(t, []string{"8.19.21"}, got)
}

func TestListElasticsearchTagsPaginatesGitHubFixture(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/repos/elastic/elasticsearch/tags", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/api/v3/repos/elastic/elasticsearch/tags?page=2>; rel="next"`, fixtureServerURL(r)))
			_, _ = w.Write([]byte(`[{"name":"v8.19.17"},{"name":"v7.17.28"}]`))
		case "2":
			_, _ = w.Write([]byte(`[{"name":"v8.19.21"},{"name":"v9.5.3"}]`))
		default:
			http.Error(w, "unexpected page", http.StatusBadRequest)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := github.NewClient(github.WithEnterpriseURLs(server.URL, server.URL))

	got, err := ListElasticsearchTags(context.Background(), client)
	require.NoError(t, err)
	assert.Equal(t, []string{"v8.19.17", "v7.17.28", "v8.19.21", "v9.5.3"}, got)
}

func fixtureServerURL(r *http.Request) string {
	return "http://" + r.Host
}
