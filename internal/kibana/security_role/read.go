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

	"github.com/hashicorp/terraform-plugin-framework/attr"

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
	updated, rd := reconcileSDKLegacyOptionalSets(updated, prior)
	diags.Append(rd...)
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

// reconcileSDKLegacyOptionalSets preserves Plugin SDK v2 state quirks: optional sets that were
// stored as known-empty (elements length 0, non-null) stay that way after Read, matching
// flatten's SetNull when the API omits the field. Without this, SDK→PF migrations show a
// non-empty plan (PlanOnly upgrade tests).
func reconcileSDKLegacyOptionalSets(out, prior resourceModel) (resourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var d diag.Diagnostics
	out.Elasticsearch, d = reconcileSingleBlockOptionalSets(
		prior.Elasticsearch,
		out.Elasticsearch,
		elasticsearchBlockAttrTypes(),
		elasticsearchBlockObjectType(),
		map[string]struct{}{"run_as": {}},
	)
	diags.Append(d...)
	out.Kibana, d = reconcileSingleBlockOptionalSets(
		prior.Kibana,
		out.Kibana,
		kibanaBlockAttrTypes(),
		kibanaBlockObjectType(),
		map[string]struct{}{"base": {}},
	)
	diags.Append(d...)
	return out, diags
}

func reconcileSingleBlockOptionalSets(
	priorSet, outSet types.Set,
	blockAttrTypes map[string]attr.Type,
	blockObjType types.ObjectType,
	nullToEmptyAttrNames map[string]struct{},
) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if priorSet.IsNull() || priorSet.IsUnknown() || outSet.IsNull() || outSet.IsUnknown() {
		return outSet, diags
	}
	pElems, oElems := priorSet.Elements(), outSet.Elements()
	if len(pElems) != 1 || len(oElems) != 1 {
		return outSet, diags
	}
	pObj, pok := pElems[0].(types.Object)
	oObj, ook := oElems[0].(types.Object)
	if !pok || !ook {
		return outSet, diags
	}
	pAttrs := pObj.Attributes()
	outAttrs := make(map[string]attr.Value, len(oObj.Attributes()))
	for k, v := range oObj.Attributes() {
		outAttrs[k] = v
	}
	changed := false
	for name := range nullToEmptyAttrNames {
		oVal, ok := outAttrs[name].(types.Set)
		if !ok || !oVal.IsNull() {
			continue
		}
		pVal, ok := pAttrs[name].(types.Set)
		if !ok || pVal.IsNull() || pVal.IsUnknown() || len(pVal.Elements()) != 0 {
			continue
		}
		outAttrs[name] = pVal
		changed = true
	}
	if !changed {
		return outSet, diags
	}
	newObj, d := types.ObjectValue(blockAttrTypes, outAttrs)
	diags.Append(d...)
	if diags.HasError() {
		return outSet, diags
	}
	newSet, d := types.SetValue(blockObjType, []attr.Value{newObj})
	diags.Append(d...)
	if diags.HasError() {
		return outSet, diags
	}
	return newSet, diags
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
