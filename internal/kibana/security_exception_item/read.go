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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readExceptionItem(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model ExceptionItemModel) (ExceptionItemModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	model.SpaceID = types.StringValue(spaceID)

	readResp, d := kibanaoapi.GetExceptionListItem(ctx, oapiClient, spaceID, exceptionItemReadParams(resourceID, model.NamespaceType))
	diags.Append(d...)
	if diags.HasError() {
		return model, false, diags
	}

	// If namespace_type was not known (e.g., during import) and the item was not found,
	// try reading with namespace_type=agnostic
	if readResp == nil && !typeutils.IsKnown(model.NamespaceType) {
		agnosticNsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
		id := resourceID
		retryParams := &kbapi.ReadExceptionListItemParams{
			Id:            &id,
			NamespaceType: &agnosticNsType,
		}
		readResp, d = kibanaoapi.GetExceptionListItem(ctx, oapiClient, spaceID, retryParams)
		diags.Append(d...)
		if diags.HasError() {
			return model, false, diags
		}
	}

	if readResp == nil {
		return model, false, diags
	}

	// Update model with response using model method
	diags = model.fromAPI(ctx, readResp)
	return model, true, diags
}

func exceptionItemReadParams(itemID string, namespaceType types.String) *kbapi.ReadExceptionListItemParams {
	id := itemID
	params := &kbapi.ReadExceptionListItemParams{Id: &id}
	if typeutils.IsKnown(namespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(namespaceType.ValueString())
		params.NamespaceType = &nsType
	}
	return params
}
