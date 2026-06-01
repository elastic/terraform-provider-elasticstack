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

package security_entity_store_entity_link

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/agentbuilder"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateEntityLink(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[entityLinkModel]) (entitycore.KibanaWriteResult[entityLinkModel], diag.Diagnostics) {
	plan := req.Plan
	prior := *req.Prior
	var diags diag.Diagnostics

	_, d := client.EnforceMinVersion(ctx, minKibanaEntityStoreResolutionVersion)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
	}

	planEntityIDs, d := agentbuilder.SetToStrings(ctx, plan.EntityIDs)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
	}

	for _, id := range planEntityIDs {
		if id == plan.TargetID.ValueString() {
			diags.AddError(
				"Self-link not allowed",
				fmt.Sprintf("target_id %q must not appear in entity_ids", plan.TargetID.ValueString()),
			)
			return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
		}
	}

	priorEntityIDs, d := agentbuilder.SetToStrings(ctx, prior.EntityIDs)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
	}

	added, removed := computeSetDiff(priorEntityIDs, planEntityIDs)

	if len(added) > 0 {
		body := kbapi.PostSecurityEntityStoreResolutionLinkJSONRequestBody{
			TargetId:  plan.TargetID.ValueString(),
			EntityIds: added,
		}
		resp, err := client.GetKibanaOapiClient().API.PostSecurityEntityStoreResolutionLinkWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(req.SpaceID))
		if err != nil {
			diags.AddError("Failed to link entities", err.Error())
			return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
		}
		if resp.StatusCode() != http.StatusOK {
			diags.Append(diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)...)
			return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
		}
	}

	if len(removed) > 0 {
		body := kbapi.PostSecurityEntityStoreResolutionUnlinkJSONRequestBody{
			EntityIds: removed,
		}
		resp, err := client.GetKibanaOapiClient().API.PostSecurityEntityStoreResolutionUnlinkWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(req.SpaceID))
		if err != nil {
			diags.AddError("Failed to unlink entities", err.Error())
			return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
		}
		if resp.StatusCode() != http.StatusOK {
			diags.Append(diagutil.ReportUnknownHTTPError(resp.StatusCode(), resp.Body)...)
			return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
		}
	}

	return entitycore.KibanaWriteResult[entityLinkModel]{Model: plan}, diags
}

func computeSetDiff(old, next []string) (added, removed []string) {
	oldSet := make(map[string]struct{}, len(old))
	for _, id := range old {
		oldSet[id] = struct{}{}
	}
	nextSet := make(map[string]struct{}, len(next))
	for _, id := range next {
		nextSet[id] = struct{}{}
	}
	for _, id := range next {
		if _, ok := oldSet[id]; !ok {
			added = append(added, id)
		}
	}
	for _, id := range old {
		if _, ok := nextSet[id]; !ok {
			removed = append(removed, id)
		}
	}
	return
}
