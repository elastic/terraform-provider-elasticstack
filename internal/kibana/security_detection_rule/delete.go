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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteDetectionRule(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, _ Data) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the rule using kbapi client
	kbClient := client.GetKibanaOapiClient()

	// Delete the rule
	uid, err := uuid.Parse(resourceID)
	if err != nil {
		diags.AddError("ID was not a valid UUID", err.Error())
		return diags
	}
	params := &kbapi.DeleteRuleParams{
		Id: &uid,
	}

	response, err := kbClient.API.DeleteRuleWithResponse(ctx, spaceID, params)
	if err != nil {
		diags.AddError(
			"Error deleting security detection rule",
			"Could not delete security detection rule: "+err.Error(),
		)
		return diags
	}

	if response.StatusCode() == 404 {
		// Rule was already deleted, which is fine
		return diags
	}

	if response.StatusCode() != 200 {
		diags.AddError(
			"Error deleting security detection rule",
			fmt.Sprintf("API returned status %d: %s", response.StatusCode(), string(response.Body)),
		)
		return diags
	}

	return diags
}
