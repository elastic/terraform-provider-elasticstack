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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var createExceptionItem = entitycore.SimpleKibanaCreate[ExceptionItemModel, kbapi.CreateExceptionListItemJSONRequestBody, kbapi.SecurityExceptionsAPIExceptionListItem](
	func(plan ExceptionItemModel, ctx context.Context) (kbapi.CreateExceptionListItemJSONRequestBody, diag.Diagnostics) {
		body, diags := plan.toCreateRequest(ctx)
		if diags.HasError() {
			return kbapi.CreateExceptionListItemJSONRequestBody{}, diags
		}
		return *body, diags
	},
	kibanaoapi.CreateExceptionListItem,
	(*ExceptionItemModel).populateCreated,
)

// populateCreated captures the item ID assigned by the create response.
func (m *ExceptionItemModel) populateCreated(_ context.Context, spaceID string, created *kbapi.SecurityExceptionsAPIExceptionListItem) diag.Diagnostics {
	var diags diag.Diagnostics
	if entitycore.RequireNonNilKibanaWriteResponse(&diags, created, "create", "exception item") {
		return diags
	}

	m.ItemID = typeutils.StringishValue(created.ItemId)
	m.ID = entitycore.KibanaResourceID(spaceID, created.Id)

	return diags
}
