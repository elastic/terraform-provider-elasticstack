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
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func fetchRole(ctx context.Context, client *clients.KibanaScopedClient, name string) (*kibanaoapi.SecurityRole, bool, diag.Diagnostics) {
	oapiClient, getDiags := client.GetKibanaOapiClient()
	if getDiags.HasError() {
		return nil, false, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to get Kibana OpenAPI client", getDiags[0].Summary()),
		}
	}
	role, apiDiags := kibanaoapi.GetSecurityRole(ctx, oapiClient, name)
	if apiDiags.HasError() {
		return nil, false, apiDiags
	}
	if role == nil {
		return nil, false, nil
	}
	return role, true, nil
}

// roleFields holds the fields that the resource and data source models share
// after flattening an API role. Metadata is a pointer: nil signals that the
// API returned no metadata, letting callers decide whether to overwrite their
// model (the resource leaves prior state untouched; the data source nulls).
type roleFields struct {
	Description   types.String
	Metadata      *jsontypes.Normalized
	Elasticsearch types.Object
	Kibana        types.Set
}

// roleFieldsFromAPI flattens a Kibana API role into the resource/data-source
// model fields. See [alignSetRepresentation] for how `hint` (plan on
// Create/Update, prior state on Read, zero-value on the data source) is
// consulted to preserve null vs known-empty representation.
func roleFieldsFromAPI(ctx context.Context, role *kibanaoapi.SecurityRole, hint roleHint) (roleFields, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := roleFields{Description: types.StringNull()}

	if role.Description != nil {
		out.Description = types.StringValue(*role.Description)
	}

	esObj, d := flattenElasticsearchObject(ctx, &role.Elasticsearch, hint.elasticsearch)
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Elasticsearch = esObj

	kibSet, d := flattenKibana(ctx, role.Kibana, hint.kibana)
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Kibana = kibSet

	if role.Metadata != nil {
		meta, md := metadataFromAPI(role)
		diags.Append(md...)
		out.Metadata = &meta
	}

	return out, diags
}

// roleHint bundles the plan/state values that flatten consults for null vs
// known-empty representation. Zero-valued members fall back to default
// flatten behaviour (null when empty).
type roleHint struct {
	elasticsearch types.Object
	kibana        types.Set
}

func hintFromResourceModel(m resourceModel) roleHint {
	return roleHint{elasticsearch: m.Elasticsearch, kibana: m.Kibana}
}

// readRoleResource is the envelope-driven resource Read entry point. The
// `prior` argument is the framework-supplied prior state and is used both as
// the canonical base for fields the API does not return (notably metadata
// when absent) and as the representation hint for flatten.
func readRoleResource(ctx context.Context, client *clients.KibanaScopedClient, resourceID, _ string, prior resourceModel) (resourceModel, bool, diag.Diagnostics) {
	return readRoleResourceWithHint(ctx, client, resourceID, prior, hintFromResourceModel(prior))
}

// readRoleResourceWithHint performs the read-after-write refresh used by
// Create and Update. `base` supplies any fields the API does not return
// (typically the prior state on Update or the plan on Create); `hint`
// supplies the representational guidance flatten uses to align null vs
// known-empty sets with the plan/state the framework will compare against.
func readRoleResourceWithHint(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, base resourceModel, hint roleHint) (resourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	role, found, d := fetchRole(ctx, client, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return base, false, diags
	}
	if !found {
		return base, false, nil
	}

	fields, d := roleFieldsFromAPI(ctx, role, hint)
	diags.Append(d...)
	if diags.HasError() {
		return base, false, diags
	}

	out := base
	out.Name = types.StringValue(resourceID)
	out.ID = types.StringValue(resourceID)
	out.Description = fields.Description
	out.Elasticsearch = fields.Elasticsearch
	out.Kibana = fields.Kibana
	if fields.Metadata != nil {
		out.Metadata = *fields.Metadata
	}
	return out, true, diags
}

func readRoleDataSource(ctx context.Context, client *clients.KibanaScopedClient, config dataSourceModel) (dataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	role, found, d := fetchRole(ctx, client, config.Name.ValueString())
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}
	if !found {
		config.Description = types.StringNull()
		config.Metadata = jsontypes.NewNormalizedNull()
		config.Elasticsearch = types.ObjectNull(elasticsearchResourceAttrTypes())
		config.Kibana = types.SetNull(kibanaBlockObjectType())
		return config, diags
	}

	fields, d := roleFieldsFromAPI(ctx, role, roleHint{})
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}

	config.Description = fields.Description
	config.Elasticsearch = fields.Elasticsearch
	config.Kibana = fields.Kibana
	if fields.Metadata != nil {
		config.Metadata = *fields.Metadata
	} else {
		config.Metadata = jsontypes.NewNormalizedNull()
	}
	return config, diags
}
