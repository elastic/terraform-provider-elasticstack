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

package ephemeral

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwephemeral "github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Terraform schema attribute keys for the ephemeral API key resource. They are
// shared between the schema definition here and the table-driven validation
// tests so the linter sees a single source of truth.
const (
	attrName            = "name"
	attrType            = "type"
	attrRoleDescriptors = "role_descriptors"
	attrAccess          = "access"
)

type tfModel struct {
	entitycore.ElasticsearchConnectionField
	KeyID               types.String                                                              `tfsdk:"key_id"`
	Name                types.String                                                              `tfsdk:"name"`
	Type                types.String                                                              `tfsdk:"type"`
	RoleDescriptors     customtypes.JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor] `tfsdk:"role_descriptors"`
	Expiration          types.String                                                              `tfsdk:"expiration"`
	ExpirationTimestamp types.Int64                                                               `tfsdk:"expiration_timestamp"`
	Metadata            jsontypes.Normalized                                                      `tfsdk:"metadata"`
	Access              types.Object                                                              `tfsdk:"access"`
	InvalidateOnClose   types.Bool                                                                `tfsdk:"invalidate_on_close"`
	APIKey              types.String                                                              `tfsdk:"api_key"`
	Encoded             types.String                                                              `tfsdk:"encoded"`
}

type closeState struct {
	KeyID             string `json:"key_id"`
	InvalidateOnClose bool   `json:"invalidate_on_close"`
}

// deleteAPIKeyFn is overridable in tests.
var deleteAPIKeyFn = elasticsearch.DeleteAPIKey

func NewResource() fwephemeral.EphemeralResource {
	return entitycore.NewElasticsearchEphemeralResource[tfModel, closeState](
		"security_api_key",
		entitycore.ElasticsearchEphemeralOptions[tfModel, closeState]{
			Schema: getSchema,
			Open:   openAPIKey,
			Close:  closeAPIKey,
		},
	)
}

// ephemeralExpirationDescription overrides the shared ExpirationDescription
// because the ephemeral flavor cross-references invalidate_on_close.
const ephemeralExpirationDescription = apikey.ExpirationDescription + " Strongly recommended when invalidate_on_close is false."

func getSchema(_ context.Context) eschema.Schema {
	return eschema.Schema{
		Description:         resourceDescription,
		MarkdownDescription: resourceDescription,
		Attributes: map[string]eschema.Attribute{
			attrName: eschema.StringAttribute{
				Description: apikey.NameDescription,
				Required:    true,
				Validators:  apikey.NameValidators(),
			},
			attrType: eschema.StringAttribute{
				Description: apikey.TypeDescription,
				Optional:    true,
				Validators:  apikey.TypeValidators(),
			},
			attrRoleDescriptors: eschema.StringAttribute{
				Description: apikey.RoleDescriptorsDescription,
				CustomType:  apikey.RoleDescriptorsCustomType(),
				Optional:    true,
				Validators:  apikey.RoleDescriptorsValidators(),
			},
			"expiration": eschema.StringAttribute{
				Description: ephemeralExpirationDescription,
				Optional:    true,
			},
			"metadata": eschema.StringAttribute{
				Description: apikey.MetadataDescription,
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			attrAccess: eschema.SingleNestedAttribute{
				Description: apikey.AccessDescription,
				Optional:    true,
				Validators:  apikey.AccessValidators(),
				Attributes:  apikey.AccessAttributesEphemeral(),
			},
			"invalidate_on_close": eschema.BoolAttribute{
				Description: "When true, invalidates the API key after the Terraform run completes. Defaults to false.",
				Optional:    true,
			},
			"key_id": eschema.StringAttribute{
				Description: apikey.KeyIDDescription,
				Computed:    true,
			},
			"api_key": eschema.StringAttribute{
				Description: apikey.APIKeyDescription,
				Sensitive:   true,
				Computed:    true,
			},
			"encoded": eschema.StringAttribute{
				Description: apikey.EncodedDescription,
				Sensitive:   true,
				Computed:    true,
			},
			"expiration_timestamp": eschema.Int64Attribute{
				Description: apikey.ExpirationTimestampDescription,
				Computed:    true,
			},
		},
	}
}

func openAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.OpenRequest[tfModel]) (entitycore.OpenResult[tfModel, closeState], diag.Diagnostics) {
	model := req.Config
	var diags diag.Diagnostics

	if effectiveAPIKeyType(model.Type).ValueString() == apikey.CrossClusterAPIKeyType {
		diags.Append(openCrossClusterAPIKey(ctx, client, &model)...)
	} else {
		diags.Append(openRESTAPIKey(ctx, client, &model)...)
	}
	if diags.HasError() {
		return entitycore.OpenResult[tfModel, closeState]{}, diags
	}

	return entitycore.OpenResult[tfModel, closeState]{
		Model: model,
		CloseState: closeState{
			KeyID:             model.KeyID.ValueString(),
			InvalidateOnClose: invalidateOnCloseValue(model.InvalidateOnClose),
		},
	}, diags
}

func closeAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.CloseRequest[closeState]) (entitycore.CloseResponse, diag.Diagnostics) {
	if !req.State.InvalidateOnClose || req.State.KeyID == "" {
		return entitycore.CloseResponse{}, nil
	}
	return entitycore.CloseResponse{}, deleteAPIKeyFn(ctx, client, req.State.KeyID)
}

func effectiveAPIKeyType(apiKeyType types.String) types.String {
	if typeutils.IsKnown(apiKeyType) && apiKeyType.ValueString() != "" {
		return apiKeyType
	}
	return basetypes.NewStringValue(apikey.DefaultAPIKeyType)
}

func invalidateOnCloseValue(value types.Bool) bool {
	if !typeutils.IsKnown(value) || value.IsNull() {
		return false
	}
	return value.ValueBool()
}

func (m tfModel) toShared() apikey.TfModel {
	return apikey.TfModel{
		ElasticsearchConnection: m.ElasticsearchConnection,
		KeyID:                   m.KeyID,
		Name:                    m.Name,
		Type:                    effectiveAPIKeyType(m.Type),
		RoleDescriptors:         m.RoleDescriptors,
		Expiration:              m.Expiration,
		ExpirationTimestamp:     m.ExpirationTimestamp,
		Metadata:                m.Metadata,
		Access:                  m.Access,
		APIKey:                  m.APIKey,
		Encoded:                 m.Encoded,
	}
}

// fromShared copies the shared fields back from an apikey.TfModel into this
// ephemeral model. The connection-block list and InvalidateOnClose are
// ephemeral-only and not touched.
func (m *tfModel) fromShared(t apikey.TfModel) {
	m.KeyID = t.KeyID
	m.Name = t.Name
	// Intentionally do not copy t.Type: the shared model normalizes unset
	// type to the default ("rest"), but `type` is Optional (non-computed) on
	// the ephemeral schema, so the framework would reject a planned value
	// that differs from the user's config. Preserve the original m.Type.
	m.RoleDescriptors = t.RoleDescriptors
	m.Expiration = t.Expiration
	m.ExpirationTimestamp = t.ExpirationTimestamp
	m.Metadata = t.Metadata
	m.Access = t.Access
	m.APIKey = t.APIKey
	m.Encoded = t.Encoded
}

func openRESTAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, model *tfModel) diag.Diagnostics {
	shared := model.toShared()
	diags := apikey.CreateRESTAPIKeyOperation(ctx, client, &shared)
	if diags.HasError() {
		return diags
	}
	model.fromShared(shared)
	return diags
}

func openCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, model *tfModel) diag.Diagnostics {
	shared := model.toShared()
	diags := apikey.CreateCrossClusterAPIKeyOperation(ctx, client, &shared)
	if diags.HasError() {
		return diags
	}
	model.fromShared(shared)
	return diags
}
