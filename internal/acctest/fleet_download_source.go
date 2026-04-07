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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/agentdownloadsource"
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

	supported, verDiags := client.EnforceMinVersion(ctx, agentdownloadsource.MinVersionFleetAgentDownloadSource)
	if verDiags.HasError() {
		t.Fatal(diagutil.SdkDiagsAsError(verDiags))
	}
	if !supported {
		return
	}

	fc, err := client.GetFleetClient()
	if err != nil {
		t.Fatal(err)
	}

	listResp, diags := fleet.ListAgentDownloadSources(ctx, fc, "")
	if diags.HasError() {
		t.Fatal(diagutil.FwDiagsAsError(diags))
	}
	if listResp != nil && listResp.JSON200 != nil {
		for _, item := range listResp.JSON200.Items {
			if item.IsDefault != nil && *item.IsDefault && item.Host != "" {
				return
			}
		}
	}

	isDefault := true
	id := fleetAcceptanceDefaultDownloadSourceID
	body := kbapi.PostFleetAgentDownloadSourcesJSONRequestBody{
		Host:      "https://artifacts.elastic.co/downloads/elastic-agent",
		Name:      "Terraform Acceptance Default Agent Download Source",
		IsDefault: &isDefault,
		Id:        &id,
	}

	_, createDiags := fleet.CreateAgentDownloadSource(ctx, fc, "", body)
	if !createDiags.HasError() {
		return
	}

	getResp, getDiags := fleet.GetAgentDownloadSource(ctx, fc, fleetAcceptanceDefaultDownloadSourceID, "")
	if getDiags.HasError() {
		t.Fatal(diagutil.FwDiagsAsError(createDiags))
	}
	if getResp == nil || getResp.StatusCode() != http.StatusOK || getResp.JSON200 == nil {
		t.Fatal(diagutil.FwDiagsAsError(createDiags))
	}

	updateBody := kbapi.PutFleetAgentDownloadSourcesSourceidJSONRequestBody{
		Host:      body.Host,
		Name:      body.Name,
		IsDefault: &isDefault,
	}
	_, updDiags := fleet.UpdateAgentDownloadSource(ctx, fc, fleetAcceptanceDefaultDownloadSourceID, "", updateBody)
	if updDiags.HasError() {
		t.Fatal(diagutil.FwDiagsAsError(updDiags))
	}
}
