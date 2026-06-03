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

package securitydetectionrule

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateDetectionRule(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[Data],
) (entitycore.KibanaWriteResult[Data], diag.Diagnostics) {
	data := req.Plan
	var diags diag.Diagnostics

	// Get the rule using kbapi client
	kbClient := client.GetKibanaOapiClient()

	// Build the update request
	updateProps, d := data.toUpdateProps(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	// Update the rule
	response, err := kbClient.API.UpdateRuleWithResponse(ctx, data.SpaceID.ValueString(), updateProps)
	if err != nil {
		diags.AddError(
			"Error updating security detection rule",
			"Could not update security detection rule: "+err.Error(),
		)
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	if response.StatusCode() != 200 {
		diags.AddError(
			"Error updating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	return entitycore.KibanaWriteResult[Data]{Model: data}, diags
}
