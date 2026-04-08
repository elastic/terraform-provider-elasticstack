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

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// upsert sends a PUT /api/streams/{name} request and reads the result back
// into a new model. It is shared by Create and Update.
func (r *Resource) upsert(ctx context.Context, planModel streamModel, diags *diag.Diagnostics) *streamModel {
	kibanaClient, err := r.client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana client", err.Error())
		return nil
	}

	spaceID := planModel.SpaceID.ValueString()
	name := planModel.Name.ValueString()

	apiReq := planModel.toAPIUpsertRequest(ctx, diags)
	if diags.HasError() {
		return nil
	}

	_, upsertDiags := kibanaoapi.UpsertStream(ctx, kibanaClient, spaceID, name, apiReq)
	diags.Append(upsertDiags...)
	if diags.HasError() {
		return nil
	}

	readModel, readDiags := r.read(ctx, planModel)
	diags.Append(readDiags...)
	if diags.HasError() {
		return nil
	}

	return readModel
}
