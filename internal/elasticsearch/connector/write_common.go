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

package connector

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func applyConnectorFanOut(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	connectorID string,
	plan, config ContentConnectorData,
	prior *ContentConnectorData,
	private entitycore.PrivateStateStorage,
	isUpdate bool,
) diag.Diagnostics {
	var diags diag.Diagnostics

	priorPipeline := fwtypes.ObjectNull(pipelineModelAttrTypes())
	priorScheduling := fwtypes.ObjectNull(schedulingModelAttrTypes())
	priorFeatures := fwtypes.ObjectNull(featuresModelAttrTypes())
	if prior != nil {
		priorPipeline = prior.Pipeline
		priorScheduling = prior.Scheduling
		priorFeatures = prior.Features
	}

	if planObjectSet(plan.Pipeline) && (!isUpdate || !skipAspectOnUpdate(plan.Pipeline, priorPipeline)) {
		pipeline := plan.toPipelineAPI(ctx, &diags)
		if diags.HasError() {
			return diags
		}
		diags.Append(esclient.UpdateConnectorPipeline(ctx, client, connectorID, pipeline)...)
	}

	if planObjectSet(plan.Scheduling) && (!isUpdate || !skipAspectOnUpdate(plan.Scheduling, priorScheduling)) {
		scheduling := plan.toSchedulingAPI(ctx, &diags)
		if diags.HasError() {
			return diags
		}
		diags.Append(esclient.UpdateConnectorScheduling(ctx, client, connectorID, scheduling)...)
	}

	if planObjectSet(plan.Features) && (!isUpdate || !skipAspectOnUpdate(plan.Features, priorFeatures)) {
		features := plan.toFeaturesAPI(ctx, &diags)
		if diags.HasError() {
			return diags
		}
		diags.Append(esclient.UpdateConnectorFeatures(ctx, client, connectorID, features)...)
	}

	apiKeySet := typeutils.IsKnown(plan.APIKeyID) || typeutils.IsKnown(plan.APIKeySecretID)
	if apiKeySet && (!isUpdate || apiKeyChanged(plan, prior)) {
		diags.Append(esclient.UpdateConnectorAPIKeyID(
			ctx,
			client,
			connectorID,
			typeutils.OptionalString(plan.APIKeyID),
			typeutils.OptionalString(plan.APIKeySecretID),
		)...)
	}

	if planMapSet(plan.ConfigurationValues) {
		planMap := configurationValuesFromModel(ctx, plan.ConfigurationValues, &diags)
		if diags.HasError() {
			return diags
		}
		configMap := configurationValuesFromModel(ctx, config.ConfigurationValues, &diags)
		if diags.HasError() {
			return diags
		}

		if preflightDiags := configurationSchemaPreflight(ctx, client, connectorID, plan.ServiceType.ValueString()); preflightDiags.HasError() {
			diags.Append(preflightDiags...)
			return diags
		}

		values := encodeConfigurationValuesWire(planMap, configMap, &diags)
		if diags.HasError() {
			return diags
		}
		diags.Append(esclient.UpdateConnectorConfiguration(ctx, client, connectorID, values)...)
		if diags.HasError() {
			return diags
		}

		storeSecretHashes(ctx, private, configMap, &diags)
		if isUpdate {
			var priorMap map[string]ConfigurationValueModel
			if prior != nil {
				priorMap = configurationValuesFromModel(ctx, prior.ConfigurationValues, &diags)
			}
			clearRemovedSecretHashes(ctx, private, priorMap, configMap, &diags)
		}
	}

	return diags
}

func apiKeyChanged(plan ContentConnectorData, prior *ContentConnectorData) bool {
	if prior == nil {
		return true
	}
	return !plan.APIKeyID.Equal(prior.APIKeyID) || !plan.APIKeySecretID.Equal(prior.APIKeySecretID)
}

func configurationSchemaPreflight(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	connectorID, serviceType string,
) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, getDiags := esclient.GetConnector(ctx, client, connectorID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return diags
	}
	if resp == nil || len(resp.Configuration) == 0 {
		diags.AddError(
			configurationSchemaNotRegisteredTitle,
			configurationSchemaNotRegisteredDetail(serviceType),
		)
	}
	return diags
}

func applyEnvelopePartialsOnUpdate(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	connectorID string,
	plan, prior ContentConnectorData,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if !plan.Name.Equal(prior.Name) || !plan.Description.Equal(prior.Description) {
		diags.Append(esclient.UpdateConnectorName(
			ctx,
			client,
			connectorID,
			typeutils.OptionalString(plan.Name),
			typeutils.OptionalString(plan.Description),
		)...)
	}
	if !plan.IndexName.Equal(prior.IndexName) {
		diags.Append(esclient.UpdateConnectorIndexName(ctx, client, connectorID, typeutils.OptionalString(plan.IndexName))...)
	}
	if !plan.ServiceType.Equal(prior.ServiceType) {
		diags.Append(esclient.UpdateConnectorServiceType(ctx, client, connectorID, plan.ServiceType.ValueString())...)
	}
	if !plan.IsNative.Equal(prior.IsNative) {
		diags.Append(esclient.UpdateConnectorNative(ctx, client, connectorID, plan.IsNative.ValueBool())...)
	}

	return diags
}
