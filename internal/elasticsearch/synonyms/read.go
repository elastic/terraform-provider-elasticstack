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

package synonyms

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readSynonymSet(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data SynonymSetData) (SynonymSetData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	rules, getDiags := elasticsearch.GetSynonymSet(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return data, false, diags
	}

	if rules == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Synonym set "%s" not found, removing from state`, resourceID))
		return data, false, diags
	}

	data.populateFromAPI(ctx, rules, &diags)
	if diags.HasError() {
		return data, false, diags
	}

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return data, false, diags
	}

	data.ID = types.StringValue(id.String())

	return data, true, diags
}
