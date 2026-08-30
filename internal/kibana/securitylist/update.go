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

package securitylist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var updateSecurityList = entitycore.SimpleKibanaUpdate[Model, kbapi.UpdateListJSONRequestBody, kbapi.SecurityListsAPIList](
	func(plan Model, _ context.Context) (kbapi.UpdateListJSONRequestBody, diag.Diagnostics) {
		req, diags := plan.toUpdateRequest()
		if diags.HasError() {
			return kbapi.UpdateListJSONRequestBody{}, diags
		}
		return *req, diags
	},
	// UpdateList takes the resource ID via the request body (see
	// Model.toUpdateRequest), not as a separate parameter, so the writeID
	// argument required by SimpleKibanaUpdate's apiUpdate shape is unused here.
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, _ string, body kbapi.UpdateListJSONRequestBody) (*kbapi.SecurityListsAPIList, diag.Diagnostics) {
		return kibanaoapi.UpdateList(ctx, client, spaceID, body)
	},
	(*Model).populateUpdated,
)

// populateUpdated only validates the update response is non-nil: the plan
// already carries the correct field values for an update.
func (m *Model) populateUpdated(_ context.Context, _ string, updated *kbapi.SecurityListsAPIList) diag.Diagnostics {
	var diags diag.Diagnostics
	entitycore.RequireNonNilKibanaWriteResponse(&diags, updated, "update", "security list")
	return diags
}
