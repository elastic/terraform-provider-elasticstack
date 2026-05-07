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

package componenttemplate

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// readComponentTemplate is the envelope read callback.
// Returns (model, true, nil) when found, (_, false, nil) when not found.
func readComponentTemplate(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, resourceID)
	if sdkDiags != nil {
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if diags.HasError() {
			return state, false, diags
		}
	}

	if tpl == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Component template "%s" not found`, resourceID))
		return state, false, diags
	}

	result, d := flattenToData(ctx, tpl, state)
	diags.Append(d...)
	if diags.HasError() {
		return state, false, diags
	}

	return result, true, diags
}
