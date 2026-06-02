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
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func deleteTrainedModelAlias(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ TFModel) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	alias := resourceID
	if alias == "" {
		diags.AddError("Invalid resource ID", "model_alias cannot be empty")
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting ML trained model alias: %s", alias))

	diags.Append(elasticsearch.DeleteMLTrainedModelAlias(ctx, client, alias)...)
	if diags.HasError() {
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted ML trained model alias: %s", alias))
	return diags
}
