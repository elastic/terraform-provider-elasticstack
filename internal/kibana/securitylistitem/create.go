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

package securitylistitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var createSecurityListItem = entitycore.SimpleKibanaCreate[Model, kbapi.CreateListItemJSONRequestBody, kbapi.SecurityListsAPIListItem](
	func(plan Model, ctx context.Context) (kbapi.CreateListItemJSONRequestBody, diag.Diagnostics) {
		req, diags := plan.toAPICreateModel(ctx)
		if diags.HasError() {
			return kbapi.CreateListItemJSONRequestBody{}, diags
		}
		return *req, diags
	},
	kibanaoapi.CreateListItem,
	(*Model).populateCreated,
)

// populateCreated captures the list item ID assigned by the create response.
func (m *Model) populateCreated(_ context.Context, spaceID string, created *kbapi.SecurityListsAPIListItem) diag.Diagnostics {
	var diags diag.Diagnostics
	if entitycore.RequireNonNilKibanaWriteResponse(&diags, created, "create", "security list item") {
		return diags
	}

	m.ListItemID = typeutils.StringishValue(created.Id)
	m.ID = entitycore.KibanaResourceID(spaceID, created.Id)

	return diags
}
