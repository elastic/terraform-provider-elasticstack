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

package acctest

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/licensetype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
)

const (
	defaultElasticsearchTransportPort = "9300"
	ccrRemoteConnectTimeout           = 45 * time.Second
	ccrRemoteConnectPollInterval      = 2 * time.Second
)

// CCRRemoteClusterAlias is the remote cluster alias registered for self-remote
// CCR tests. It is unique per test process so that CCR test packages running
// concurrently (gotestsum runs package binaries in parallel) do not clobber one
// another's shared persistent remote-cluster setting.
var CCRRemoteClusterAlias = fmt.Sprintf("acc-ccr-remote-%d", os.Getpid())

// CCRTestEnv holds identifiers shared by CCR acceptance tests.
type CCRTestEnv struct {
	RemoteClusterAlias string
	RemoteProxyAddress string
}

// PreCheckCCR runs standard acceptance pre-checks, skips when CCR is unavailable, registers
// the self-remote cluster in proxy mode using the Elasticsearch transport endpoint, and returns
// configuration shared by CCR acceptance tests. Proxy-mode remote clusters open raw transport
// (TCP) connections; for a self-remote the ES node connects to proxy_address from inside its
// own container, so the address must be host:transport_port (not the HTTP port).
func PreCheckCCR(t *testing.T) CCRTestEnv {
	t.Helper()
	PreCheck(t)
	versionutils.SkipIfUnsupported(t, nil, versionutils.FlavorStateful)
	SkipIfCCRUnavailable(t)

	proxyAddress, err := elasticsearchProxyAddressFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	preCheckCCRRemote(t, proxyAddress)

	return CCRTestEnv{
		RemoteClusterAlias: CCRRemoteClusterAlias,
		RemoteProxyAddress: proxyAddress,
	}
}

// SkipIfCCRUnavailable skips the test when the cluster license does not include CCR.
func SkipIfCCRUnavailable(t *testing.T) {
	t.Helper()
	SkipIfNotAcceptanceTest(t)

	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.GetESClient().License.Get().Do(ctx)
	if err != nil {
		t.Fatalf("failed to get Elasticsearch license: %v", err)
	}

	switch res.License.Type {
	case licensetype.Trial, licensetype.Platinum, licensetype.Enterprise:
		return
	default:
		t.Skipf(
			"Skipping CCR acceptance test: license type %q does not include CCR (need trial, platinum, or enterprise)",
			res.License.Type,
		)
	}
}

func preCheckCCRRemote(t *testing.T, proxyAddress string) {
	t.Helper()

	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatal(err)
	}

	if err := registerCCRRemoteCluster(ctx, client, CCRRemoteClusterAlias, proxyAddress); err != nil {
		t.Fatalf("failed to register CCR self-remote cluster %q: %v", CCRRemoteClusterAlias, err)
	}

	connected, err := waitForRemoteClusterConnected(ctx, client, CCRRemoteClusterAlias, ccrRemoteConnectTimeout)
	if err != nil {
		t.Fatalf("failed to poll remote cluster %q connection status: %v", CCRRemoteClusterAlias, err)
	}
	if !connected {
		t.Skipf(
			"Skipping CCR acceptance test: self-remote cluster %q did not connect to transport endpoint %q within %s (proxy mode requires a reachable transport port)",
			CCRRemoteClusterAlias,
			proxyAddress,
			ccrRemoteConnectTimeout,
		)
	}
}

func registerCCRRemoteCluster(ctx context.Context, client *clients.ElasticsearchScopedClient, alias, proxyAddress string) error {
	settings := map[string]json.RawMessage{
		fmt.Sprintf("cluster.remote.%s.mode", alias):          json.RawMessage("\"proxy\""),
		fmt.Sprintf("cluster.remote.%s.proxy_address", alias): json.RawMessage(fmt.Sprintf("%q", proxyAddress)),
	}

	_, err := client.GetESClient().Cluster.PutSettings().Persistent(settings).Do(ctx)
	return err
}

func waitForRemoteClusterConnected(ctx context.Context, client *clients.ElasticsearchScopedClient, alias string, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.GetESClient().Cluster.RemoteInfo().Do(ctx)
		if err != nil {
			return false, err
		}

		if remoteClusterConnected(resp[alias]) {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(ccrRemoteConnectPollInterval):
		}
	}

	return false, nil
}

func remoteClusterConnected(info types.ClusterRemoteInfo) bool {
	if info == nil {
		return false
	}

	switch v := info.(type) {
	case types.ClusterRemoteProxyInfo:
		return v.Connected
	case *types.ClusterRemoteProxyInfo:
		return v != nil && v.Connected
	case types.ClusterRemoteSniffInfo:
		return v.Connected
	case *types.ClusterRemoteSniffInfo:
		return v != nil && v.Connected
	default:
		data, err := json.Marshal(info)
		if err != nil {
			return false
		}

		var connected struct {
			Connected bool `json:"connected"`
		}
		if err := json.Unmarshal(data, &connected); err != nil {
			return false
		}
		return connected.Connected
	}
}

func elasticsearchProxyAddressFromEnv() (string, error) {
	if override := strings.TrimSpace(os.Getenv("ELASTICSEARCH_REMOTE_PROXY_ADDRESS")); override != "" {
		return override, nil
	}

	raw := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	if raw == "" {
		return "", fmt.Errorf("ELASTICSEARCH_ENDPOINTS is not set")
	}

	endpoint := strings.TrimSpace(strings.Split(raw, ",")[0])
	if endpoint == "" {
		return "", fmt.Errorf("ELASTICSEARCH_ENDPOINTS is empty")
	}

	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse ELASTICSEARCH_ENDPOINTS %q: %w", endpoint, err)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("ELASTICSEARCH_ENDPOINTS %q has no host", os.Getenv("ELASTICSEARCH_ENDPOINTS"))
	}

	host, _, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		host = parsed.Host
	}

	transportPort := strings.TrimSpace(os.Getenv("ELASTICSEARCH_TRANSPORT_PORT"))
	if transportPort == "" {
		transportPort = defaultElasticsearchTransportPort
	}

	return net.JoinHostPort(host, transportPort), nil
}
