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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	clientconfig "github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
)

var Providers map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	providerServerFactory, err := provider.ProtoV6ProviderServerFactory(context.Background(), provider.AccTestVersion)
	if err != nil {
		log.Fatal(err)
	}
	Providers = map[string]func() (tfprotov6.ProviderServer, error){
		"elasticstack": func() (tfprotov6.ProviderServer, error) {
			server := providerServerFactory()
			if server == nil {
				return nil, fmt.Errorf("provider server factory returned nil")
			}
			return server, nil
		},
	}
}

func PreCheck(t *testing.T) {
	_, elasticsearchEndpointsOk := os.LookupEnv("ELASTICSEARCH_ENDPOINTS")
	_, kibanaEndpointOk := os.LookupEnv("KIBANA_ENDPOINT")
	_, userOk := os.LookupEnv("ELASTICSEARCH_USERNAME")
	_, passOk := os.LookupEnv("ELASTICSEARCH_PASSWORD")
	_, apiKeyOk := os.LookupEnv("ELASTICSEARCH_API_KEY")
	_, kbUserOk := os.LookupEnv("KIBANA_USERNAME")
	_, kbPassOk := os.LookupEnv("KIBANA_PASSWORD")
	_, kbAPIKeyOk := os.LookupEnv("KIBANA_API_KEY")

	if !elasticsearchEndpointsOk {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must be set for acceptance tests to run")
	}

	if !kibanaEndpointOk {
		t.Fatal("KIBANA_ENDPOINT must be set for acceptance tests to run")
	}

	authOk := (userOk && passOk) || (kbUserOk && kbPassOk) || apiKeyOk || kbAPIKeyOk
	if !authOk {
		t.Fatal("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD, or KIBANA_USERNAME and KIBANA_PASSWORD, or ELASTICSEARCH_API_KEY, or KIBANA_API_KEY must be set for acceptance tests to run")
	}
}

func PreCheckWithExplicitKibanaEndpoint(t *testing.T) {
	t.Setenv(clientconfig.PreferConfiguredKibanaEndpointEnvVar, "true")
}

// PreCheckWithWorkflowsEnabled runs standard pre-checks, skips if the server
// version is below minVersion, and ensures the workflows UI setting is enabled
// in Kibana.
func PreCheckWithWorkflowsEnabled(t *testing.T, minVersion *version.Version) {
	PreCheck(t)

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	serverVersion, diags := client.ServerVersion(context.Background())
	if diags.HasError() {
		t.Fatalf("Failed to get server version: %v", diags)
	}
	if serverVersion.LessThan(minVersion) {
		t.Skipf("Skipping test: server version %s is below minimum %s", serverVersion, minVersion)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("Failed to get Kibana client: %v", err)
	}

	// Try the internal settings API endpoint
	settingsURL := fmt.Sprintf("%s/internal/kibana/settings/workflows:ui:enabled", kibanaClient.URL)
	body := map[string]any{
		"value": true,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", settingsURL, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("x-elastic-internal-origin", "Kibana")

	resp, err := kibanaClient.HTTP.Do(req)
	if err != nil {
		t.Fatalf("Failed to enable workflows: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to enable workflows (status %d): %s. Make sure workflows are enabled in kibana.yml with 'xpack.aiAssistant.workflows.enabled: true'", resp.StatusCode, string(respBody))
	}
}

func NamedTestCaseDirectory(name string) config.TestStepConfigFunc {
	return func(tscr config.TestStepConfigRequest) string {
		return path.Join(config.TestNameDirectory()(tscr), name)
	}
}
