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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentImageForTwoDigitMinorUsesElasticRegistry(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "docker.elastic.co/elastic-agent/elastic-agent:8.10.4", AgentImage("8.10.4"))
}

func TestAgentImageFor81UsesDockerHub(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "elastic/elastic-agent:8.1.3", AgentImage("8.1.3"))
}

func TestAgentImageFor80UsesDockerHub(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "elastic/elastic-agent:8.0.1", AgentImage("8.0.1"))
}

func TestComposeStackImagesForModernMinor(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{
		"docker.elastic.co/elasticsearch/elasticsearch:8.19.21",
		"docker.elastic.co/kibana/kibana:8.19.21",
		"docker.elastic.co/elastic-agent/elastic-agent:8.19.21",
	}, ComposeStackImages("8.19.21"))
}

func TestComposeStackImagesFor81UsesDockerHubAgent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{
		"docker.elastic.co/elasticsearch/elasticsearch:8.1.3",
		"docker.elastic.co/kibana/kibana:8.1.3",
		"elastic/elastic-agent:8.1.3",
	}, ComposeStackImages("8.1.3"))
}

func TestProbeComposeStackTrueWhenAllManifestsResolve(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/elasticsearch/elasticsearch/manifests/8.19.21",
			"/v2/kibana/kibana/manifests/8.19.21",
			"/v2/elastic-agent/elastic-agent/manifests/8.19.21":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.19.21")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestProbeComposeStackFalseWhenAnyManifestMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/kibana/kibana/manifests/8.19.21" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.19.21")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestProbeComposeStackRoutes81AgentToDockerHub(t *testing.T) {
	t.Parallel()

	var hubAgentHits int
	elastic := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "elastic-agent") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(elastic.Close)

	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/elastic/elastic-agent/manifests/8.1.3" {
			hubAgentHits++
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(hub.Close)

	prober := RegistryProber{
		Client:            elastic.Client(),
		ElasticRegistry:   elastic.URL,
		DockerHubRegistry: hub.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.1.3")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, hubAgentHits)
}

func TestProbeComposeStackAuthenticatesThenFindsElasticImages(t *testing.T) {
	t.Parallel()

	server := newRegistryAuthServer(t, map[string]int{
		"/v2/elasticsearch/elasticsearch/manifests/8.19.21": http.StatusOK,
		"/v2/kibana/kibana/manifests/8.19.21":               http.StatusOK,
		"/v2/elastic-agent/elastic-agent/manifests/8.19.21": http.StatusOK,
	})

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.19.21")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestProbeComposeStackAuthenticatesThenReportsMissingElasticImage(t *testing.T) {
	t.Parallel()

	server := newRegistryAuthServer(t, map[string]int{
		"/v2/elasticsearch/elasticsearch/manifests/8.19.21": http.StatusOK,
		"/v2/kibana/kibana/manifests/8.19.21":               http.StatusNotFound,
		"/v2/elastic-agent/elastic-agent/manifests/8.19.21": http.StatusOK,
	})

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.19.21")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestProbeComposeStackAuthenticatesThenFindsDockerHubAgent(t *testing.T) {
	t.Parallel()

	server := newRegistryAuthServer(t, map[string]int{
		"/v2/elasticsearch/elasticsearch/manifests/8.1.3": http.StatusOK,
		"/v2/kibana/kibana/manifests/8.1.3":               http.StatusOK,
		"/v2/elastic/elastic-agent/manifests/8.1.3":       http.StatusOK,
	})

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.1.3")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestProbeComposeStackAuthenticatesThenReportsMissingDockerHubAgent(t *testing.T) {
	t.Parallel()

	server := newRegistryAuthServer(t, map[string]int{
		"/v2/elasticsearch/elasticsearch/manifests/8.1.3": http.StatusOK,
		"/v2/kibana/kibana/manifests/8.1.3":               http.StatusOK,
		"/v2/elastic/elastic-agent/manifests/8.1.3":       http.StatusNotFound,
	})

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.1.3")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestProbeComposeStackErrorsOn5xxAfterAuth(t *testing.T) {
	t.Parallel()

	server := newRegistryAuthServer(t, map[string]int{
		"/v2/elasticsearch/elasticsearch/manifests/8.19.21": http.StatusOK,
		"/v2/kibana/kibana/manifests/8.19.21":               http.StatusServiceUnavailable,
		"/v2/elastic-agent/elastic-agent/manifests/8.19.21": http.StatusOK,
	})

	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	ok, err := ProbeComposeStack(context.Background(), prober, "8.19.21")
	require.Error(t, err)
	assert.False(t, ok)
}

func TestProbeComposeStackErrorsOnTransportFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	prober := RegistryProber{
		Client:            server.Client(),
		ElasticRegistry:   server.URL,
		DockerHubRegistry: server.URL,
	}
	server.Close()

	ok, err := ProbeComposeStack(context.Background(), prober, "8.19.21")
	require.Error(t, err)
	assert.False(t, ok)
}

func newRegistryAuthServer(t *testing.T, manifestStatus map[string]int) *httptest.Server {
	t.Helper()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"token":"anonymous-pull-token"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer anonymous-pull-token" {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="token-service",scope="repository:test:pull"`, server.URL))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if status, ok := manifestStatus[r.URL.Path]; ok {
			w.WriteHeader(status)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	return server
}
