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

package apikey

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createcrossclusterapikey"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ ephemeral.EphemeralResource               = (*EphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure    = (*EphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithClose        = (*EphemeralResource)(nil)
)

const ephemeralPrivateDataKey = "elasticstack.security_api_key"

type EphemeralResource struct {
	client *clients.ProviderClientFactory
}

type ephemeralTfModel struct {
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

type elasticsearchConnectionObject struct {
	Username               types.String `tfsdk:"username"`
	Password               types.String `tfsdk:"password"`
	APIKey                 types.String `tfsdk:"api_key"`
	BearerToken            types.String `tfsdk:"bearer_token"`
	ESClientAuthentication types.String `tfsdk:"es_client_authentication"`
	Endpoints              types.List   `tfsdk:"endpoints"`
	Headers                types.Map    `tfsdk:"headers"`
	Insecure               types.Bool   `tfsdk:"insecure"`
	CAFile                 types.String `tfsdk:"ca_file"`
	CAData                 types.String `tfsdk:"ca_data"`
	CertFile               types.String `tfsdk:"cert_file"`
	CertData               types.String `tfsdk:"cert_data"`
	KeyFile                types.String `tfsdk:"key_file"`
	KeyData                types.String `tfsdk:"key_data"`
}

func NewEphemeralResource() ephemeral.EphemeralResource {
	return &EphemeralResource{}
}

func (r *EphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_elasticsearch_security_api_key", req.ProviderTypeName)
}

func (r *EphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = factory
}

func (r *EphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = getEphemeralSchema()
}

func getEphemeralSchema() eschema.Schema {
	return eschema.Schema{
		Description:         ephemeralResourceDescription,
		MarkdownDescription: ephemeralResourceDescription,
		Blocks: map[string]eschema.Block{
			"elasticsearch_connection": providerschema.GetEsEphemeralConnectionBlock(),
		},
		Attributes: map[string]eschema.Attribute{
			"name": eschema.StringAttribute{
				Description: "Specifies the name for this API key.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([[:graph:]]| )+$`),
						apiKeyNameInvalidMessage,
					),
				},
			},
			"type": eschema.StringAttribute{
				Description: "The type of API key. Valid values are 'rest' (default) and 'cross_cluster'. Cross-cluster API keys are used for cross-cluster search and replication.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(defaultAPIKeyType, crossClusterAPIKeyType),
				},
			},
			"role_descriptors": eschema.StringAttribute{
				Description: "Role descriptors for this API key.",
				CustomType:  customtypes.NewJSONWithDefaultsType(populateRoleDescriptorsDefaults),
				Optional:    true,
				Validators: []validator.String{
					requiresType(defaultAPIKeyType),
				},
			},
			"expiration": eschema.StringAttribute{
				Description: "Expiration time for the API key. By default, API keys never expire. Strongly recommended when invalidate_on_close is false.",
				Optional:    true,
			},
			"metadata": eschema.StringAttribute{
				Description: "Arbitrary metadata that you want to associate with the API key.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"access": eschema.SingleNestedAttribute{
				Description: "Access configuration for cross-cluster API keys. Only applicable when type is 'cross_cluster'.",
				Optional:    true,
				Validators: []validator.Object{
					requiresType(crossClusterAPIKeyType),
				},
				Attributes: map[string]eschema.Attribute{
					"search": eschema.ListNestedAttribute{
						Description: "A list of search configurations for which the cross-cluster API key will have search privileges.",
						Optional:    true,
						NestedObject: eschema.NestedAttributeObject{
							Attributes: map[string]eschema.Attribute{
								"names": eschema.ListAttribute{
									Description: "A list of index patterns for search.",
									Required:    true,
									ElementType: types.StringType,
								},
								"field_security": eschema.StringAttribute{
									Description: "Field-level security configuration in JSON format.",
									Optional:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								"query": eschema.StringAttribute{
									Description: "Query to filter documents for search operations in JSON format.",
									Optional:    true,
									CustomType:  jsontypes.NormalizedType{},
								},
								"allow_restricted_indices": eschema.BoolAttribute{
									Description: "Whether to allow access to restricted indices.",
									Optional:    true,
								},
							},
						},
					},
					"replication": eschema.ListNestedAttribute{
						Description: "A list of replication configurations for which the cross-cluster API key will have replication privileges.",
						Optional:    true,
						NestedObject: eschema.NestedAttributeObject{
							Attributes: map[string]eschema.Attribute{
								"names": eschema.ListAttribute{
									Description: "A list of index patterns for replication.",
									Required:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
			"invalidate_on_close": eschema.BoolAttribute{
				Description: "When true, invalidates the API key after the Terraform run completes. Defaults to false.",
				Optional:    true,
			},
			"key_id": eschema.StringAttribute{
				Description: "Unique id for this API key.",
				Computed:    true,
			},
			"api_key": eschema.StringAttribute{
				Description: "Generated API Key.",
				Sensitive:   true,
				Computed:    true,
			},
			"encoded": eschema.StringAttribute{
				Description: "API key credentials which is the Base64-encoding of the UTF-8 representation of the id and api_key joined by a colon (:).",
				Sensitive:   true,
				Computed:    true,
			},
			"expiration_timestamp": eschema.Int64Attribute{
				Description: "Expiration time in milliseconds for the API key. By default, API keys never expire.",
				Computed:    true,
			},
		},
	}
}

func (r *EphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Ephemeral Resource",
			"The ephemeral resource was not configured. Configure must run before Open.",
		)
		return
	}

	var model ephemeralTfModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Type = effectiveAPIKeyType(model.Type)

	client, clientDiags := r.client.GetElasticsearchClient(ctx, model.ElasticsearchConnection)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.Type.ValueString() == crossClusterAPIKeyType {
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

func (r *EphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
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
	return basetypes.NewStringValue(defaultAPIKeyType)
}

func invalidateOnCloseValue(value types.Bool) bool {
	if !typeutils.IsKnown(value) || value.IsNull() {
		return false
	}
	return value.ValueBool()
}

func (m ephemeralTfModel) toTfModel() tfModel {
	return tfModel{
		ElasticsearchConnection: m.ElasticsearchConnection,
		KeyID:                   m.KeyID,
		Name:                    m.Name,
		Type:                    m.Type,
		RoleDescriptors:         m.RoleDescriptors,
		Expiration:              m.Expiration,
		ExpirationTimestamp:     m.ExpirationTimestamp,
		Metadata:                m.Metadata,
		Access:                  m.Access,
		APIKey:                  m.APIKey,
		Encoded:                 m.Encoded,
	}
}

func (m *ephemeralTfModel) populateFromCreate(apiKey *createapikey.Response) {
	m.KeyID = basetypes.NewStringValue(apiKey.Id)
	m.Name = basetypes.NewStringValue(apiKey.Name)
	m.APIKey = basetypes.NewStringValue(apiKey.ApiKey)
	m.Encoded = basetypes.NewStringValue(apiKey.Encoded)
	m.ExpirationTimestamp = basetypes.NewInt64Value(0)
	if apiKey.Expiration != nil && *apiKey.Expiration > 0 {
		m.ExpirationTimestamp = basetypes.NewInt64Value(*apiKey.Expiration)
	}
}

func (m *ephemeralTfModel) populateFromCrossClusterCreate(apiKey *createcrossclusterapikey.Response) {
	m.KeyID = basetypes.NewStringValue(apiKey.Id)
	m.Name = basetypes.NewStringValue(apiKey.Name)
	m.APIKey = basetypes.NewStringValue(apiKey.ApiKey)
	m.Encoded = basetypes.NewStringValue(apiKey.Encoded)
	m.ExpirationTimestamp = basetypes.NewInt64Value(0)
	if apiKey.Expiration != nil && *apiKey.Expiration > 0 {
		m.ExpirationTimestamp = basetypes.NewInt64Value(*apiKey.Expiration)
	}
}

func (r *EphemeralResource) openRESTAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, model *ephemeralTfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	tfModel := model.toTfModel()
	diags.Append(validateRestrictionSupport(ctx, client, tfModel)...)
	if diags.HasError() {
		return diags
	}

	createRequest, modelDiags := tfModel.toAPICreateRequest()
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	putResponse, createDiags := elasticsearch.CreateAPIKey(ctx, client, createRequest)
	diags.Append(createDiags...)
	if diags.HasError() {
		return diags
	}
	if putResponse == nil {
		diags.AddError("API Key Creation Failed", "API key creation returned nil response")
		return diags
	}

	model.populateFromCreate(putResponse)
	return diags
}

func (r *EphemeralResource) openCrossClusterAPIKey(ctx context.Context, client *clients.ElasticsearchScopedClient, model *ephemeralTfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	tfModel := model.toTfModel()
	diags.Append(entitycore.EnforceVersionRequirements(ctx, client, &tfModel)...)
	if diags.HasError() {
		return diags
	}

	createRequest, modelDiags := tfModel.toCrossClusterAPICreateRequest(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		return diags
	}

	putResponse, createDiags := elasticsearch.CreateCrossClusterAPIKey(ctx, client, createRequest)
	diags.Append(createDiags...)
	if diags.HasError() {
		return diags
	}
	if putResponse == nil {
		diags.AddError("API Key Creation Failed", "Cross-cluster API key creation returned nil response")
		return diags
	}

	model.populateFromCrossClusterCreate(putResponse)
	return diags
}

func closeAPIKeyIfRequested(
	ctx context.Context,
	factory *clients.ProviderClientFactory,
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

	return elasticsearch.DeleteAPIKey(ctx, client, keyID)
}

func saveEphemeralPrivateData(ctx context.Context, privateState ephemeralPrivateState, model ephemeralTfModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if privateState == nil {
		return diags
	}

	connectionJSON := ""
	if typeutils.IsKnown(model.ElasticsearchConnection) && !model.ElasticsearchConnection.IsNull() {
		encodedConnection, encodeDiags := encodeElasticsearchConnection(ctx, model.ElasticsearchConnection)
		diags.Append(encodeDiags...)
		if diags.HasError() {
			return diags
		}
		connectionJSON = encodedConnection
	}

	payload, err := json.Marshal(ephemeralPrivateData{
		KeyID:             model.KeyID.ValueString(),
		InvalidateOnClose: invalidateOnCloseValue(model.InvalidateOnClose),
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

func encodeElasticsearchConnection(ctx context.Context, connection types.List) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(connection) || connection.IsNull() {
		return "", diags
	}

	var connectionObjects []elasticsearchConnectionObject
	diags.Append(connection.ElementsAs(ctx, &connectionObjects, false)...)
	if diags.HasError() {
		return "", diags
	}

	bytes, err := json.Marshal(connectionObjects)
	if err != nil {
		diags.AddError("Failed to marshal elasticsearch_connection for Close", err.Error())
		return "", diags
	}

	return string(bytes), diags
}

func decodeElasticsearchConnection(ctx context.Context, connectionJSON string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if connectionJSON == "" {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	var connectionObjects []elasticsearchConnectionObject
	if err := json.Unmarshal([]byte(connectionJSON), &connectionObjects); err != nil {
		diags.AddError("Failed to parse elasticsearch_connection from ephemeral private data", err.Error())
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	objectValues := make([]attr.Value, 0, len(connectionObjects))
	for _, connectionObject := range connectionObjects {
		objectValue, objectDiags := types.ObjectValueFrom(
			ctx,
			providerschema.ElasticsearchConnectionObjectType().AttrTypes,
			map[string]attr.Value{
				"username":                 connectionObject.Username,
				"password":                 connectionObject.Password,
				"api_key":                  connectionObject.APIKey,
				"bearer_token":             connectionObject.BearerToken,
				"es_client_authentication": connectionObject.ESClientAuthentication,
				"endpoints":                connectionObject.Endpoints,
				"headers":                  connectionObject.Headers,
				"insecure":                 connectionObject.Insecure,
				"ca_file":                  connectionObject.CAFile,
				"ca_data":                  connectionObject.CAData,
				"cert_file":                connectionObject.CertFile,
				"cert_data":                connectionObject.CertData,
				"key_file":                 connectionObject.KeyFile,
				"key_data":                 connectionObject.KeyData,
			},
		)
		diags.Append(objectDiags...)
		if diags.HasError() {
			return providerschema.ElasticsearchConnectionNullList(), diags
		}
		objectValues = append(objectValues, objectValue)
	}

	connection, listDiags := types.ListValue(providerschema.ElasticsearchConnectionObjectType(), objectValues)
	diags.Append(listDiags...)
	if diags.HasError() {
		return providerschema.ElasticsearchConnectionNullList(), diags
	}

	return connection, diags
}
