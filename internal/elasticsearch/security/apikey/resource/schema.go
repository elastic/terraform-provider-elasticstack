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

package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/planmodifiers"
)

// getSchema returns the schema for the given version. The elasticsearch_connection
// block is omitted; the ElasticsearchResource envelope injects it.
func getSchema(version int64) schema.Schema {
	return schema.Schema{
		Version:     version,
		Description: resourceDescription,
		Blocks:      map[string]schema.Block{},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_id": schema.StringAttribute{
				Description: apikey.KeyIDDescription,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: apikey.NameDescription,
				Required:    true,
				Validators:  apikey.NameValidators(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrType: schema.StringAttribute{
				Description: apikey.TypeDescription,
				Optional:    true,
				Computed:    true,
				Validators:  apikey.TypeValidators(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
					planmodifiers.StringUseDefaultIfUnknown(apikey.DefaultAPIKeyType),
				},
			},
			attrRoleDescriptors: schema.StringAttribute{
				Description: apikey.RoleDescriptorsDescription,
				CustomType:  apikey.RoleDescriptorsCustomType(),
				Optional:    true,
				Computed:    true,
				Validators:  apikey.RoleDescriptorsValidators(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					requiresReplaceIfUpdateNotSupported(),
					SetUnknownIfAccessHasChanges(),
				},
			},
			attrExpiration: schema.StringAttribute{
				Description: apikey.ExpirationDescription,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration_timestamp": schema.Int64Attribute{
				Description: apikey.ExpirationTimestampDescription,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			attrMetadata: schema.StringAttribute{
				Description: apikey.MetadataDescription,
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					requiresReplaceIfUpdateNotSupported(),
				},
			},
			attrAccess: schema.SingleNestedAttribute{
				Description: apikey.AccessDescription,
				Optional:    true,
				Validators:  apikey.AccessValidators(),
				Attributes:  apikey.AccessAttributesResource(),
			},
			"api_key": schema.StringAttribute{
				Description: apikey.APIKeyDescription,
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"encoded": schema.StringAttribute{
				Description: apikey.EncodedDescription,
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func requiresReplaceBecauseUpdateNotSupported(ctx context.Context, priv privateData) (bool, diag.Diagnostics) {
	caps, diags := apikeyCapabilitiesOfLastRead(ctx, priv)
	if diags.HasError() {
		return false, diags
	}
	return caps != nil && !caps.SupportsUpdate, diags
}

func requiresReplaceIfUpdateNotSupported() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, res planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			requiresReplace, readDiags := requiresReplaceBecauseUpdateNotSupported(ctx, res.Private)
			resp.Diagnostics.Append(readDiags...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.RequiresReplace = requiresReplace
		},
		"Requires replace if the server does not support update",
		"Requires replace if the server does not support update",
	)
}
