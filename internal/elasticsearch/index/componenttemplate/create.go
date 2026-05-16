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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeComponentTemplate is the envelope Create/Update callback. It calls
// PutComponentTemplate, sets the composite ID during create, and returns the
// model. The envelope invokes readComponentTemplate after this callback to
// refresh state.
func writeComponentTemplate(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.WriteRequest[Data]) (entitycore.WriteResult[Data], diag.Diagnostics) {
	var diags diag.Diagnostics
	plan := req.Plan

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}
	diags.Append(datastreamoptions.EnforceMinServerVersion(plan.Template, serverVersion)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}

	componentTemplate, d := expandFromData(ctx, plan)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutComponentTemplate(ctx, client, &componentTemplate))...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}

	compositeID, idDiags := client.ID(ctx, req.WriteID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(idDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{Model: plan}, diags
	}
	plan.ID = types.StringValue(compositeID.String())

	return entitycore.WriteResult[Data]{Model: plan}, diags
}
