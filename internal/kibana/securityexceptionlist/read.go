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

package securityexceptionlist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readExceptionList(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, prior ExceptionListModel) (ExceptionListModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient, getDiags := client.GetKibanaOapiClient()
	diags.Append(getDiags...)
	if diags.HasError() {
		return prior, false, diags
	}

	prior.SpaceID = types.StringValue(spaceID)

	id := resourceID
	params := &kbapi.ReadExceptionListParams{
		Id: &id,
	}
	if typeutils.IsKnown(prior.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(prior.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}

	readResp, d := kibanaoapi.GetExceptionList(ctx, oapiClient, spaceID, params)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if readResp == nil && !typeutils.IsKnown(prior.NamespaceType) {
		agnosticNsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
		params.NamespaceType = &agnosticNsType
		readResp, d = kibanaoapi.GetExceptionList(ctx, oapiClient, spaceID, params)
		diags.Append(d...)
		if diags.HasError() {
			return prior, false, diags
		}
	}

	if readResp == nil {
		return prior, false, diags
	}

	diags.Append(prior.fromAPI(ctx, readResp)...)
	return prior, true, diags
}
