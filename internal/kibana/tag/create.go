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

package tag

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func createTag(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tagModel],
) (entitycore.KibanaWriteResult[tagModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()
	body := plan.toAPIModel(true)
	explicitTagID := typeutils.IsKnown(plan.TagID) && plan.TagID.ValueString() != ""

	if explicitTagID {
		tagID := plan.TagID.ValueString()
		existing, getDiags := kibanaoapi.GetTag(ctx, oapiClient, req.SpaceID, tagID)
		diags.Append(getDiags...)
		if diags.HasError() {
			return entitycore.KibanaWriteResult[tagModel]{}, diags
		}

		switch tagCreateActionForExisting(true, existing != nil) {
		case tagCreateActionRejectExisting:
			diags.AddError(
				"Tag already exists",
				fmt.Sprintf("A tag with ID %q already exists in space %q. Import it with: terraform import elasticstack_kibana_tag.<name> '%s/%s'",
					tagID, req.SpaceID, req.SpaceID, tagID),
			)
			return entitycore.KibanaWriteResult[tagModel]{}, diags
		case tagCreateActionPUT:
			detail, upsertDiags := kibanaoapi.UpsertTag(ctx, oapiClient, req.SpaceID, tagID, body)
			diags.Append(upsertDiags...)
			if diags.HasError() {
				return entitycore.KibanaWriteResult[tagModel]{}, diags
			}
			plan.setCompositeIdentity(req.SpaceID, detail.ID)
			return entitycore.KibanaWriteResult[tagModel]{Model: plan}, diags
		default:
			diags.AddError("Invalid create action", "Internal error resolving tag create strategy.")
			return entitycore.KibanaWriteResult[tagModel]{}, diags
		}
	}

	detail, createDiags := kibanaoapi.CreateTag(ctx, oapiClient, req.SpaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tagModel]{}, diags
	}

	plan.setCompositeIdentity(req.SpaceID, detail.ID)
	return entitycore.KibanaWriteResult[tagModel]{Model: plan}, diags
}
