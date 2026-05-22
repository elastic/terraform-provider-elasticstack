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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwephemeral "github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ fwephemeral.EphemeralResource              = (*Resource)(nil)
	_ fwephemeral.EphemeralResourceWithConfigure = (*Resource)(nil)
	_ fwephemeral.EphemeralResourceWithClose     = (*Resource)(nil)
)

const ephemeralPrivateDataKey = "elasticstack.security_api_key"

// Resource implements the ephemeral resource for `elasticstack_elasticsearch_security_api_key`.
type Resource struct {
	client *clients.ProviderClientFactory
}

type tfModel struct {
	ElasticsearchConnection types.List                                                                `tfsdk:"elasticsearch_connection"`
	KeyID                   types.String                                                              `tfsdk:"key_id"`
	Name                    types.String                                                              `tfsdk:"name"`
	Type                    types.String                                                              `tfsdk:"type"`
	RoleDescriptors         customtypes.JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor] `tfsdk:"role_descriptors"`
	Expiration              types.String                                                              `tfsdk:"expiration"`
	ExpirationTimestamp     types.Int64                                                               `tfsdk:"expiration_timestamp"`
	Metadata                jsontypes.Normalized                                                      `tfsdk:"metadata"`
	Access                  types.Object                                                              `tfsdk:"access"`
	InvalidateOnClose       types.Bool                                                                `tfsdk:"invalidate_on_close"`
	APIKey                  types.String                                                              `tfsdk:"api_key"`
	Encoded                 types.String                                                              `tfsdk:"encoded"`
}

type ephemeralPrivateState interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

type ephemeralPrivateData struct {
	KeyID             string `json:"key_id"`
	InvalidateOnClose bool   `json:"invalidate_on_close"`
	ConnectionJSON    string `json:"connection_json,omitempty"`
}

// deleteAPIKeyFn is overridable in tests.
var deleteAPIKeyFn = elasticsearch.DeleteAPIKey

type elasticsearchClientResolver interface {
	GetElasticsearchClient(ctx context.Context, connection types.List) (*clients.ElasticsearchScopedClient, diag.Diagnostics)
}

func NewResource() fwephemeral.EphemeralResource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req fwephemeral.MetadataRequest, resp *fwephemeral.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_elasticsearch_security_api_key", req.ProviderTypeName)
}

func (r *Resource) Configure(_ context.Context, req fwephemeral.ConfigureRequest, resp *fwephemeral.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = factory
}

func (r *Resource) Schema(_ context.Context, _ fwephemeral.SchemaRequest, resp *fwephemeral.SchemaResponse) {
	resp.Schema = getSchema()
}

// ephemeralExpirationDescription overrides the shared ExpirationDescription
// because the ephemeral flavor cross-references invalidate_on_close.
const ephemeralExpirationDescription = apikey.ExpirationDescription + " Strongly recommended when invalidate_on_close is false."

func getSchema() eschema.Schema {
	return eschema.Schema{
		Description:         resourceDescription,
		MarkdownDescription: resourceDescription,
		Blocks: map[string]eschema.Block{
			"elasticsearch_connection": providerschema.GetEsEphemeralConnectionBlock(),
		},
		Attributes: map[string]eschema.Attribute{
			"name": eschema.StringAttribute{
				Description: apikey.NameDescription,
				Required:    true,
				Validators:  apikey.NameValidators(),
			},
			"type": eschema.StringAttribute{
				Description: apikey.TypeDescription,
				Optional:    true,
				Validators:  apikey.TypeValidators(),
			},
			"role_descriptors": eschema.StringAttribute{
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
			"access": eschema.SingleNestedAttribute{
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

func (r *Resource) Open(ctx context.Context, req fwephemeral.OpenRequest, resp *fwephemeral.OpenResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Ephemeral Resource",
			"The ephemeral resource was not configured. Configure must run before Open.",
		)
		return
	}

	var model tfModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := r.client.GetElasticsearchClient(ctx, model.ElasticsearchConnection)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if effectiveAPIKeyType(model.Type).ValueString() == apikey.CrossClusterAPIKeyType {
		resp.Diagnostics.Append(r.openCrossClusterAPIKey(ctx, client, &model)...)
	} else {
		resp.Diagnostics.Append(r.openRESTAPIKey(ctx, client, &model)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(saveEphemeralPrivateData(ctx, resp.Private, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &model)...)
}

func (r *Resource) Close(ctx context.Context, req fwephemeral.CloseRequest, resp *fwephemeral.CloseResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Ephemeral Resource",
			"The ephemeral resource was not configured. Configure must run before Close.",
		)
		return
	}

	privateData, diags := loadEphemeralPrivateData(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if privateData == nil {
		return
	}

	connection, connDiags := elasticsearchConnectionFromPrivateJSON(ctx, privateData.ConnectionJSON)
	resp.Diagnostics.Append(connDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(closeAPIKeyIfRequested(ctx, r.client, connection, privateData.InvalidateOnClose, privateData.KeyID)...)
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
	m.Type = t.Type
	m.RoleDescriptors = t.RoleDescriptors
	m.Expiration = t.Expiration
	m.ExpirationTimestamp = t.ExpirationTimestamp
	m.Metadata = t.Metadata
	m.Access = t.Access
	m.APIKey = t.APIKey
	m.Encoded = t.Encoded
}

func (r *Resource) openRESTAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, model *tfModel) diag.Diagnostics {
	shared := model.toShared()
	diags := apikey.CreateRESTAPIKeyOperation(ctx, client, &shared)
	if diags.HasError() {
		return diags
	}
	model.fromShared(shared)
	return diags
}

func (r *Resource) openCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, model *tfModel) diag.Diagnostics {
	shared := model.toShared()
	diags := apikey.CreateCrossClusterAPIKeyOperation(ctx, client, &shared)
	if diags.HasError() {
		return diags
	}
	model.fromShared(shared)
	return diags
}

func closeAPIKeyIfRequested(
	ctx context.Context,
	factory elasticsearchClientResolver,
	connection types.List,
	invalidateOnClose bool,
	keyID string,
) diag.Diagnostics {
	if !invalidateOnClose || keyID == "" {
		return nil
	}

	client, diags := factory.GetElasticsearchClient(ctx, connection)
	if diags.HasError() {
		return diags
	}

	return deleteAPIKeyFn(ctx, client, keyID)
}

func saveEphemeralPrivateData(ctx context.Context, privateState ephemeralPrivateState, model tfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	// Close() is a no-op unless invalidate_on_close is true, so skip persisting
	// private state (including the connection snapshot) in the common case.
	if privateState == nil || !invalidateOnCloseValue(model.InvalidateOnClose) {
		return diags
	}

	connectionJSON, encodeDiags := encodeElasticsearchConnection(ctx, model.ElasticsearchConnection)
	diags.Append(encodeDiags...)
	if diags.HasError() {
		return diags
	}

	payload, err := json.Marshal(ephemeralPrivateData{
		KeyID:             model.KeyID.ValueString(),
		InvalidateOnClose: true,
		ConnectionJSON:    connectionJSON,
	})
	if err != nil {
		diags.AddError("Failed to marshal ephemeral private data", err.Error())
		return diags
	}

	diags.Append(privateState.SetKey(ctx, ephemeralPrivateDataKey, payload)...)
	return diags
}

func loadEphemeralPrivateData(ctx context.Context, privateState ephemeralPrivateState) (*ephemeralPrivateData, diag.Diagnostics) {
	var diags diag.Diagnostics
	if privateState == nil {
		return nil, diags
	}

	raw, keyDiags := privateState.GetKey(ctx, ephemeralPrivateDataKey)
	diags.Append(keyDiags...)
	if diags.HasError() || len(raw) == 0 {
		return nil, diags
	}

	var data ephemeralPrivateData
	if err := json.Unmarshal(raw, &data); err != nil {
		diags.AddError("Failed to parse ephemeral private data", err.Error())
		return nil, diags
	}

	return &data, diags
}

func elasticsearchConnectionFromPrivateJSON(ctx context.Context, connectionJSON string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if connectionJSON == "" {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	return decodeElasticsearchConnection(ctx, connectionJSON)
}
