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

package parameter

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if strings.Count(request.ID, "/") > 1 {
		response.Diagnostics.AddError(
			fmt.Sprintf("Failed to parse parameter import ID %s", request.ID),
			fmt.Sprintf(
				"Import ID must use at most one slash in the form `<space_id>/<parameter_uuid>` or a bare `<parameter_uuid>`. Current value: %s",
				request.ID,
			),
		)
		return
	}

	if strings.Contains(request.ID, "/") {
		// ResolveCompositeSpaceAndID leaves malformed composite strings (for example
		// "<space_id>/" with an empty resource segment) as the full raw ID when
		// CompositeIDFromStr fails, so import would succeed with an invalid UUID.
		// Pre-check CompositeIDFromStr to surface the same rejection as state identity parsing.
		if _, diags := clients.CompositeIDFromStr(request.ID); diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	spaceID, resourceID := clients.ResolveCompositeSpaceAndID(types.StringNull(), request.ID)
	if resourceID == "" {
		response.Diagnostics.AddError(
			"Wrong resource ID.",
			"Import ID must include a parameter UUID in the form `<space_id>/<parameter_uuid>` or a bare `<parameter_uuid>`.",
		)
		return
	}

	spaceID = clients.EffectiveSpaceID(spaceID)
	compositeID := (&clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}).String()

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), compositeID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("space_id"), spaceID)...)
}
