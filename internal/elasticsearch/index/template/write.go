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

package template

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeIndexTemplate handles both Create (req.Prior == nil) and Update. It returns
// a config-derived seed model (ID and elasticsearch_connection filled in) so the
// envelope read-after-write path can call readIndexTemplate without plan Unknown
// placeholder drift.
func writeIndexTemplate(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[Model]) (entitycore.WriteResult[Model], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan

	indexTemplate, buildDiags := plan.toAPIModel(ctx)
	diags.Append(buildDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{}, diags
	}

	if req.Prior != nil {
		applyAllowCustomRouting8xWorkaround(ctx, *req.Prior, req.Config, indexTemplate)
	}

	diags.Append(elasticsearch.PutIndexTemplate(ctx, client, indexTemplate)...)
	if diags.HasError() {
		return entitycore.WriteResult[Model]{}, diags
	}

	priorForRead := req.Config
	priorForRead.ElasticsearchConnection = plan.ElasticsearchConnection

	if req.Prior == nil {
		id, sdkDiags := client.ID(ctx, plan.Name.ValueString())
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if diags.HasError() {
			return entitycore.WriteResult[Model]{}, diags
		}
		priorForRead.ID = types.StringValue(id.String())
	} else {
		priorForRead.ID = req.Prior.ID
	}

	return entitycore.WriteResult[Model]{Model: priorForRead}, diags
}
