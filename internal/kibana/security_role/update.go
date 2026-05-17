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

package security_role

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateRole(ctx context.Context, client *clients.KibanaScopedClient, resourceID, _ string, plan, prior resourceModel) (resourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	roleName, body, d := expandResourceModel(ctx, plan)
	diags.Append(d...)
	if diags.HasError() {
		return prior, diags
	}
	if roleName != resourceID {
		diags.AddError("Internal error", "resource name mismatch during update")
		return prior, diags
	}
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Unable to get Kibana OpenAPI client", err.Error())
		return prior, diags
	}
	createOnly := false
	params := kbapi.PutSecurityRoleNameParams{
		CreateOnly: &createOnly,
	}
	sdkDiags := kibanaoapi.PutSecurityRole(ctx, oapiClient, roleName, params, body)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return prior, diags
	}
	updated, _, rd := readRoleResource(ctx, client, roleName, "", prior)
	diags.Append(rd...)
	return updated, diags
}
