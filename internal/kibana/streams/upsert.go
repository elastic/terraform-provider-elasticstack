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

package streams

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// writeStream issues a PUT /api/streams/{name} request. It is shared by Create
// and Update; the envelope performs the read-after-write that refreshes state.
func writeStream(ctx context.Context, apiClient *clients.KibanaScopedClient, planModel streamModel) diag.Diagnostics {
	var diags diag.Diagnostics

	kibanaClient := apiClient.GetKibanaOapiClient()

	spaceID := planModel.GetSpaceID().ValueString()
	name := planModel.GetResourceID().ValueString()

	apiReq := planModel.toAPIUpsertRequest(ctx, &diags)
	if diags.HasError() {
		return diags
	}

	_, upsertDiags := kibanaoapi.UpsertStream(ctx, kibanaClient, spaceID, name, apiReq)
	diags.Append(upsertDiags...)
	return diags
}
