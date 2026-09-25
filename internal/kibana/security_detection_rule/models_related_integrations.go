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

package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper function to process related integrations configuration for all rule types
func (d Data) relatedIntegrationsToAPI(ctx context.Context) (*kbapi.SecurityDetectionsAPIRelatedIntegrationArray, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiRelatedIntegrations := convertListFieldToAPI(ctx, d.RelatedIntegrations, path.Root("related_integrations"), &diags,
		func(integration RelatedIntegrationModel, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPIRelatedIntegration {
			apiIntegration := kbapi.SecurityDetectionsAPIRelatedIntegration{
				Package: integration.Package.ValueString(),
				Version: integration.Version.ValueString(),
			}

			// Set optional integration field if provided
			if typeutils.IsKnown(integration.Integration) {
				integrationName := integration.Integration.ValueString()
				apiIntegration.Integration = &integrationName
			}

			return apiIntegration
		})
	if apiRelatedIntegrations == nil {
		return nil, diags
	}

	return &apiRelatedIntegrations, diags
}

// convertRelatedIntegrationsToModel converts kbapi.SecurityDetectionsAPIRelatedIntegrationArray to Terraform model
func convertRelatedIntegrationsToModel(ctx context.Context, apiRelatedIntegrations kbapi.SecurityDetectionsAPIRelatedIntegrationArray) (types.List, diag.Diagnostics) {
	return convertAPISliceToListField(ctx, apiRelatedIntegrations, getRelatedIntegrationElementType(),
		func(apiIntegration kbapi.SecurityDetectionsAPIRelatedIntegration) RelatedIntegrationModel {
			return RelatedIntegrationModel{
				Package: types.StringValue(apiIntegration.Package),
				Version: types.StringValue(apiIntegration.Version),

				// Set optional integration field if provided
				Integration: typeutils.StringishPointerValue(apiIntegration.Integration),
			}
		})
}

func (d *Data) updateRelatedIntegrationsFromAPI(ctx context.Context, relatedIntegrations *kbapi.SecurityDetectionsAPIRelatedIntegrationArray) diag.Diagnostics {
	var diags diag.Diagnostics
	var slice kbapi.SecurityDetectionsAPIRelatedIntegrationArray
	if relatedIntegrations != nil {
		slice = *relatedIntegrations
	}
	d.RelatedIntegrations, diags = updateListFieldFromAPI(ctx, slice,
		types.ListNull(getRelatedIntegrationElementType()),
		convertRelatedIntegrationsToModel)
	return diags
}
