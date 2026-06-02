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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createDetectionRule(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[Data],
) (entitycore.KibanaWriteResult[Data], diag.Diagnostics) {
	data := req.Plan
	var diags diag.Diagnostics

	// Create the rule using kbapi client
	kbClient := client.GetKibanaOapiClient()

	// Build the create request
	createProps, d := data.toCreateProps(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	// Create the rule
	response, err := kbClient.API.CreateRuleWithResponse(ctx, data.SpaceID.ValueString(), createProps)
	if err != nil {
		diags.AddError(
			"Error creating security detection rule",
			"Could not create security detection rule: "+err.Error(),
		)
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	if response.StatusCode() != 200 {
		diags.AddError(
			"Error creating security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	// Set the ID based on the created rule
	id, d := extractID(response.JSON200)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[Data]{}, diags
	}

	compID := clients.CompositeID{
		ClusterID:  data.SpaceID.ValueString(),
		ResourceID: id,
	}
	data.ID = types.StringValue(compID.String())
	data.RuleID = types.StringValue(id)

	return entitycore.KibanaWriteResult[Data]{Model: data}, diags
}
