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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func fetchRole(ctx context.Context, client *clients.KibanaScopedClient, name string) (*kibanaoapi.SecurityRole, bool, diag.Diagnostics) {
	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return nil, false, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to get Kibana OpenAPI client", err.Error()),
		}
	}
	role, sdkDiags := kibanaoapi.GetSecurityRole(ctx, oapiClient, name)
	fwDiags := diagutil.FrameworkDiagsFromSDK(sdkDiags)
	if fwDiags.HasError() {
		return nil, false, fwDiags
	}
	if role == nil {
		return nil, false, nil
	}
	return role, true, nil
}

func readRoleResource(ctx context.Context, client *clients.KibanaScopedClient, resourceID, _ string, prior resourceModel) (resourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	role, found, d := fetchRole(ctx, client, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}
	if !found {
		return prior, false, nil
	}
	updated, d := resourceModelFromAPI(ctx, role, resourceID, prior)
	diags.Append(d...)
	return updated, true, diags
}

func resourceModelFromAPI(ctx context.Context, role *kibanaoapi.SecurityRole, name string, prior resourceModel) (resourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := prior
	out.Name = types.StringValue(name)
	out.ID = types.StringValue(name)

	if role.Description != nil {
		out.Description = types.StringValue(*role.Description)
	} else {
		out.Description = types.StringNull()
	}

	esSet, d := flattenElasticsearch(ctx, &role.Elasticsearch)
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Elasticsearch = esSet

	kibSet, d := flattenKibana(ctx, role.Kibana)
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Kibana = kibSet

	if role.Metadata != nil {
		meta, md := metadataFromAPI(role)
		diags.Append(md...)
		out.Metadata = meta
	}

	return out, diags
}

func readRoleDataSource(ctx context.Context, client *clients.KibanaScopedClient, config dataSourceModel) (dataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	name := config.Name.ValueString()
	role, found, d := fetchRole(ctx, client, name)
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}
	if !found {
		config.Description = types.StringNull()
		config.Metadata = jsontypes.NewNormalizedNull()
		config.Elasticsearch = types.SetNull(elasticsearchBlockObjectType())
		config.Kibana = types.SetNull(kibanaBlockObjectType())
		return config, diags
	}
	if role.Description != nil {
		config.Description = types.StringValue(*role.Description)
	} else {
		config.Description = types.StringValue("")
	}
	if role.Metadata != nil {
		meta, md := metadataFromAPI(role)
		diags.Append(md...)
		config.Metadata = meta
	} else {
		config.Metadata = jsontypes.NewNormalizedNull()
	}
	esSet, d := flattenElasticsearch(ctx, &role.Elasticsearch)
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}
	config.Elasticsearch = esSet
	kibSet, d := flattenKibana(ctx, role.Kibana)
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}
	config.Kibana = kibSet
	return config, diags
}
