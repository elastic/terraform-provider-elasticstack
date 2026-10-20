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

package agentbuilderagent

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateAgent(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[agentModel]) (entitycore.KibanaWriteResult[agentModel], diag.Diagnostics) {
	// Whether to send is decided from config; the mode itself is read from the
	// plan, which equals config whenever config is non-null.
	// Avatar fields are Optional+Computed with UseStateForUnknown, so the plan
	// carries the prior value when config omits them; only send configured values.
	send := updateSendFlags{
		AccessControl: sendsAccessControl(req.Config.AccessControl),
		AvatarColor:   typeutils.IsKnown(req.Config.AvatarColor),
		AvatarSymbol:  typeutils.IsKnown(req.Config.AvatarSymbol),
	}
	diags := enforceAccessControlVersion(ctx, client, send.AccessControl)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	supportsSkillIDs, d := client.EnforceMinVersion(ctx, agentbuilder.MinExtendedAPIVersion)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[agentModel]{}, diags
	}

	return entitycore.SimpleKibanaUpdate[agentModel, kbapi.PutAgentBuilderAgentsIdJSONRequestBody, models.Agent](
		func(plan agentModel, ctx context.Context, _ string) (kbapi.PutAgentBuilderAgentsIdJSONRequestBody, diag.Diagnostics) {
			return plan.toAPIUpdateModel(ctx, supportsSkillIDs, send)
		},
		kibanaoapi.UpdateAgent,
		(*agentModel).setWriteSpaceID,
	)(ctx, client, req)
}
