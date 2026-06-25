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

package osquerypack

import (
	"context"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (m osqueryPackModel) toCreateRequestBody(ctx context.Context) (kbapi.OsqueryCreatePacksJSONRequestBody, diag.Diagnostics) {
	return m.toWriteRequestBody(ctx)
}

func (m osqueryPackModel) toUpdateRequestBody(ctx context.Context) (kbapi.OsqueryUpdatePacksJSONRequestBody, diag.Diagnostics) {
	createBody, diags := m.toWriteRequestBody(ctx)
	if diags.HasError() {
		return kbapi.OsqueryUpdatePacksJSONRequestBody{}, diags
	}

	return kbapi.OsqueryUpdatePacksJSONRequestBody(createBody), diags
}

func (m osqueryPackModel) toWriteRequestBody(ctx context.Context) (kbapi.SecurityOsqueryAPICreatePacksRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := kbapi.SecurityOsqueryAPICreatePacksRequestBody{}

	if typeutils.IsKnown(m.Name) {
		name := m.Name.ValueString()
		body.Name = &name
	}

	if typeutils.IsKnown(m.Description) {
		desc := m.Description.ValueString()
		body.Description = &desc
	}

	if typeutils.IsKnown(m.Enabled) {
		enabled := m.Enabled.ValueBool()
		body.Enabled = &enabled
	}

	policyIDs, d := policyIDsToAPI(ctx, m.PolicyIDs)
	diags.Append(d...)
	body.PolicyIds = policyIDs

	shards := shardsMapToAPI(m.Shards)
	body.Shards = shards

	queries, d := queriesMapToAPI(ctx, m.Queries)
	diags.Append(d...)
	body.Queries = queries

	return body, diags
}

func policyIDsToAPI(ctx context.Context, set types.Set) (*kbapi.SecurityOsqueryAPIPolicyIds, diag.Diagnostics) {
	if set.IsUnknown() || set.IsNull() {
		return nil, nil
	}

	var ids []string
	diags := set.ElementsAs(ctx, &ids, false)
	if diags.HasError() {
		return nil, diags
	}

	sort.Strings(ids)
	return &ids, diags
}

func shardsMapToAPI(shards types.Map) *kbapi.SecurityOsqueryAPIShards {
	if shards.IsUnknown() || shards.IsNull() {
		return nil
	}

	result := make(kbapi.SecurityOsqueryAPIShards, len(shards.Elements()))
	for policyID, av := range shards.Elements() {
		result[policyID] = float32(av.(types.Float64).ValueFloat64())
	}

	return &result
}

func queriesMapToAPI(ctx context.Context, queries types.Map) (*kbapi.SecurityOsqueryAPIObjectQueries, diag.Diagnostics) {
	if !typeutils.IsKnown(queries) || queries.IsNull() {
		return nil, nil
	}

	var diags diag.Diagnostics
	result := make(kbapi.SecurityOsqueryAPIObjectQueries, len(queries.Elements()))

	for name, av := range queries.Elements() {
		var q queryModel
		d := av.(basetypes.ObjectValue).As(ctx, &q, basetypes.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		item, d := q.toAPIType(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		result[name] = item
	}

	if len(result) == 0 {
		return nil, nil
	}

	return &result, diags
}
