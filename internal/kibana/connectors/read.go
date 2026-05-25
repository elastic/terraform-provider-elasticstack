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

package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readConnector(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	model tfModel,
) (tfModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, getDiags := client.GetKibanaOapiClient()
	diags.Append(getDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	connector, readDiags := kibanaoapi.GetConnector(ctx, oapiClient, resourceID, spaceID)
	if connector == nil && readDiags == nil {
		return model, false, diags
	}
	diags.Append(readDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}
	diags.Append(model.populateFromAPI(connector, compositeID)...)
	if diags.HasError() {
		return model, false, diags
	}

	return model, true, diags
}
