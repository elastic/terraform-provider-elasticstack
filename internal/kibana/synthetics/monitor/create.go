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

package monitor

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var MinLabelsVersion = version.Must(version.NewVersion("8.16.0"))

func createMonitor(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModelV0],
) (entitycore.KibanaWriteResult[tfModelV0], diag.Diagnostics) {
	planModel := req.Plan
	var diags diag.Diagnostics

	diags.Append(planModel.enforceVersionConstraints(ctx, client)...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	input, apiDiags := planModel.toKibanaAPIRequest(ctx)
	diags.Append(apiDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	oapiClient := client.GetKibanaOapiClient()

	spaceID := req.SpaceID
	result, createDiags := kibanaoapi.CreateMonitor(ctx, oapiClient, spaceID, *input)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	if result == nil {
		diags.AddError(
			fmt.Sprintf("Failed to create Kibana monitor `%s`, space %s", planModel.Name.ValueString(), spaceID),
			"empty response from API",
		)
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}

	updatedPlan, modelDiags := planModel.toModelV0(ctx, result, spaceID)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModelV0]{}, diags
	}
	planModel = *updatedPlan

	return entitycore.KibanaWriteResult[tfModelV0]{Model: planModel}, diags
}
