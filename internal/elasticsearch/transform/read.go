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

package transform

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// readTransform fetches the transform definition and stats and returns a
// populated model. It returns found=false when the transform no longer exists.
func readTransform(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, state tfModel) (tfModel, bool, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	transform, sdkDiags := elasticsearch.GetTransform(ctx, client, &resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	if transform == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Transform "%s" not found, removing from state`, resourceID))
		return state, false, diags
	}

	stats, sdkDiags := elasticsearch.GetTransformStats(ctx, client, &resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return state, false, diags
	}

	model, convDiags := fromAPIModel(ctx, transform, stats, state)
	diags.Append(convDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	return model, true, diags
}
