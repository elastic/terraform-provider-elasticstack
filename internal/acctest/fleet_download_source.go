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
	"net/http"
	"sync"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
)

// Stable ID for acceptance-test bootstrap; lives in the default Kibana space and is not managed by Terraform state.
const fleetAcceptanceDefaultDownloadSourceID = "terraform-acc-fleet-default-download-source"

var ensureFleetDefaultDownloadSourceOnce sync.Once

func ensureFleetDefaultAgentDownloadSource(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatal(err)
	}

	fc, err := client.GetFleetClient()
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: We intentionally do NOT version-gate this bootstrap.
	// Some stacks require a default download source for agent policy operations even when our
	// elasticstack_fleet_agent_download_source resource is version-gated.
	// If the endpoint isn't available, we no-op; otherwise we ensure a default exists.
	listResp, err := fc.API.GetFleetAgentDownloadSourcesWithResponse(ctx)
	if err != nil {
		t.Fatal(err)
	}
	switch listResp.StatusCode() {
	case http.StatusOK:
		for _, item := range listResp.JSON200.Items {
			if item.IsDefault != nil && *item.IsDefault && item.Host != "" {
				return
			}
		}
	case http.StatusNotFound:
		// Endpoint not available on this stack; nothing we can do here.
		return
	default:
		t.Fatal("unexpected response when listing Fleet agent download sources: " + string(listResp.Body))
	}

	isDefault := true
	id := fleetAcceptanceDefaultDownloadSourceID
	body := kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{
		Host:      "https://artifacts.elastic.co/downloads/elastic-agent",
		Name:      "Terraform Acceptance Default Agent Download Source",
		IsDefault: &isDefault,
		Id:        &id,
	}

	createResp, err := fc.API.PostFleetAgentDownloadSourcesWithResponse(ctx, body)
	if err != nil {
		t.Fatal(err)
	}
	switch createResp.StatusCode() {
	case http.StatusOK:
		return
	case http.StatusNotFound:
		// Endpoint not available on this stack; nothing we can do here.
		return
	}

	getResp, err := fc.API.GetFleetAgentDownloadSourcesSourceidWithResponse(ctx, fleetAcceptanceDefaultDownloadSourceID)
	if err != nil {
		t.Fatal(err)
	}
	if getResp.StatusCode() != http.StatusOK || getResp.JSON200 == nil {
		t.Fatal("failed to create default Fleet agent download source: " + string(createResp.Body))
	}

	updateBody := kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
		Host:      body.Host,
		Name:      body.Name,
		IsDefault: &isDefault,
	}
	updResp, err := fc.API.PutFleetAgentDownloadSourcesSourceidWithResponse(ctx, fleetAcceptanceDefaultDownloadSourceID, updateBody)
	if err != nil {
		t.Fatal(err)
	}
	switch updResp.StatusCode() {
	case http.StatusOK:
		return
	case http.StatusNotFound:
		return
	default:
		t.Fatal("failed to update default Fleet agent download source: " + string(updResp.Body))
	}
}
