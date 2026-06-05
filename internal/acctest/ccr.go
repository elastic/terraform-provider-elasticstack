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
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/licensetype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
)

// CCRRemoteClusterAlias is the remote cluster alias registered for self-remote CCR tests.
const CCRRemoteClusterAlias = "acc-ccr-remote"

// CCRTestEnv holds identifiers shared by CCR acceptance tests.
type CCRTestEnv struct {
	RemoteClusterAlias string
	RemoteProxyAddress string
}

// PreCheckCCR runs standard acceptance pre-checks, skips when CCR is unavailable, and
// returns configuration for self-remote CCR using proxy mode (HTTP endpoint of the
// test cluster). Proxy mode works from CI containers and local host runs without
// requiring the transport port to be published.
func PreCheckCCR(t *testing.T) CCRTestEnv {
	t.Helper()
	PreCheck(t)
	versionutils.SkipIfUnsupported(t, nil, versionutils.FlavorStateful)
	SkipIfCCRUnavailable(t)

	proxyAddress, err := elasticsearchProxyAddressFromEnv()
	if err != nil {
		t.Fatal(err)
	}

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

func elasticsearchProxyAddressFromEnv() (string, error) {
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

	return parsed.Host, nil
}
