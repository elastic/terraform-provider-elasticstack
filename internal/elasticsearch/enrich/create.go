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

package enrich

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// upsertEnrichPolicy handles both Create and Update; the enrich policy PUT API
// is idempotent so the same callback serves both lifecycle methods. Note: the
// schema marks every field RequiresReplace so Update only ever runs against an
// unchanged plan in practice.
func upsertEnrichPolicy(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	req entitycore.WriteRequest[PolicyDataWithExecute],
) (entitycore.WriteResult[PolicyDataWithExecute], diag.Diagnostics) {
	var diags diag.Diagnostics
	data := req.Plan
	resourceID := req.WriteID

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[PolicyDataWithExecute]{Model: data}, diags
	}

	indices := typeutils.SetTypeAs[string](ctx, data.Indices, path.Empty(), &diags)
	if diags.HasError() {
		return entitycore.WriteResult[PolicyDataWithExecute]{Model: data}, diags
	}

	enrichFields := typeutils.SetTypeAs[string](ctx, data.EnrichFields, path.Empty(), &diags)
	if diags.HasError() {
		return entitycore.WriteResult[PolicyDataWithExecute]{Model: data}, diags
	}

	policy := &models.EnrichPolicy{
		Type:         data.PolicyType.ValueString(),
		Name:         resourceID,
		Indices:      indices,
		MatchField:   data.MatchField.ValueString(),
		EnrichFields: enrichFields,
	}

	if !data.Query.IsNull() && !data.Query.IsUnknown() {
		policy.Query = data.Query.ValueString()
	}

	if sdkDiags := elasticsearch.PutEnrichPolicy(ctx, client, policy); sdkDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		return entitycore.WriteResult[PolicyDataWithExecute]{Model: data}, diags
	}

	data.ID = types.StringValue(id.String())

	if !data.Execute.IsNull() && !data.Execute.IsUnknown() && data.Execute.ValueBool() {
		if sdkDiags := elasticsearch.ExecuteEnrichPolicy(ctx, client, resourceID); sdkDiags.HasError() {
			diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
			return entitycore.WriteResult[PolicyDataWithExecute]{Model: data}, diags
		}
	}

	return entitycore.WriteResult[PolicyDataWithExecute]{Model: data}, diags
}
