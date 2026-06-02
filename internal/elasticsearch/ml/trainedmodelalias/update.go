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

package trainedmodelalias

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	elasticsearch "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func updateTrainedModelAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[TFModel]) (entitycore.WriteResult[TFModel], fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	alias := req.WriteID
	plan := req.Plan

	tflog.Debug(ctx, fmt.Sprintf("Updating ML trained model alias: %s", alias))

	modelID := plan.ModelID.ValueString()
	reassign := plan.Reassign.ValueBool()

	diags.Append(elasticsearch.PutMLTrainedModelAlias(ctx, client, modelID, alias, reassign)...)
	if diags.HasError() {
		return entitycore.WriteResult[TFModel]{Model: plan}, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully updated ML trained model alias: %s", alias))
	return entitycore.WriteResult[TFModel]{Model: plan}, diags
}
