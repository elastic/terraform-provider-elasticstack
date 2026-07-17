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

package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputModel struct {
	entitycore.ResourceTimeoutsField
	ID                          types.String                    `tfsdk:"id"`
	KibanaConnection            types.List                      `tfsdk:"kibana_connection"`
	OutputID                    types.String                    `tfsdk:"output_id"`
	Name                        types.String                    `tfsdk:"name"`
	Type                        types.String                    `tfsdk:"type"`
	Hosts                       types.List                      `tfsdk:"hosts"` // > string
	ServiceToken                types.String                    `tfsdk:"service_token"`
	CaSha256                    types.String                    `tfsdk:"ca_sha256"`
	CaTrustedFingerprint        types.String                    `tfsdk:"ca_trusted_fingerprint"`
	DefaultIntegrations         types.Bool                      `tfsdk:"default_integrations"`
	DefaultMonitoring           types.Bool                      `tfsdk:"default_monitoring"`
	ConfigYaml                  customtypes.NormalizedYamlValue `tfsdk:"config_yaml"`
	SpaceIDs                    types.Set                       `tfsdk:"space_ids"` // > string
	Ssl                         types.Object                    `tfsdk:"ssl"`       // > outputSslModel
	Kafka                       types.Object                    `tfsdk:"kafka"`     // > outputKafkaModel
	SyncIntegrations            types.Bool                      `tfsdk:"sync_integrations"`
	SyncUninstalledIntegrations types.Bool                      `tfsdk:"sync_uninstalled_integrations"`
	WriteToLogsStreams          types.Bool                      `tfsdk:"write_to_logs_streams"`
}

func (model outputModel) GetID() types.String             { return model.ID }
func (model outputModel) GetResourceID() types.String     { return model.OutputID }
func (model outputModel) GetKibanaConnection() types.List { return model.KibanaConnection }

func (model outputModel) GetSpaceID() types.String {
	if model.SpaceIDs.IsNull() || model.SpaceIDs.IsUnknown() {
		return types.StringValue("")
	}
	for _, elem := range model.SpaceIDs.Elements() {
		s, ok := elem.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			continue
		}
		if v := s.ValueString(); v != "" {
			return s
		}
	}
	return types.StringValue("")
}

// IsUnscopedSpace implements entitycore.KibanaUnscopedSpace.
func (model outputModel) IsUnscopedSpace() bool { return true }

func (model outputModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	var reqs []entitycore.VersionRequirement

	if sslModel := typeutils.ObjectTypeAs[outputSslModel](ctx, model.Ssl, path.Root("ssl"), nil); sslModel != nil {
		if typeutils.IsKnown(sslModel.VerificationMode) {
			reqs = append(reqs, entitycore.VersionRequirement{
				MinVersion:   *MinVersionOutputSSLVerificationMode,
				ErrorMessage: fmt.Sprintf("ssl.verification_mode requires server version %s or higher", MinVersionOutputSSLVerificationMode.String()),
			})
		}
	}

	if model.Type.ValueString() == outputTypeKafka {
		reqs = append(reqs, entitycore.VersionRequirement{
			MinVersion:   *MinVersionOutputKafka,
			ErrorMessage: fmt.Sprintf("Kafka output type requires server version %s or higher", MinVersionOutputKafka.String()),
		})
	}

	return reqs, nil
}

func (model *outputModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) (diags diag.Diagnostics) {
	if union == nil {
		return
	}

	output, err := union.ValueByDiscriminator()
	if err != nil {
		diags.AddError(err.Error(), "")
		return
	}

	switch output := output.(type) {
	case kbapi.KibanaHTTPAPIsOutputResponseElasticsearch:
		diags.Append(model.fromAPIElasticsearchModel(ctx, &output)...)

	case kbapi.KibanaHTTPAPIsOutputResponseLogstash:
		diags.Append(model.fromAPILogstashModel(ctx, &output)...)

	case kbapi.KibanaHTTPAPIsOutputResponseKafka:
		diags.Append(model.fromAPIKafkaModel(ctx, &output)...)
	case kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch:
		diags.Append(model.fromAPIRemoteElasticsearchModel(ctx, &output)...)
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %T", output), "")
	}

	return
}

func (model outputModel) toAPICreateModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	outputType := model.Type.ValueString()

	switch outputType {
	case outputTypeElasticsearch:
		return model.toAPICreateElasticsearchModel(ctx)
	case outputTypeLogstash:
		return model.toAPICreateLogstashModel(ctx)
	case outputTypeKafka:
		return model.toAPICreateKafkaModel(ctx)
	case outputTypeRemoteElasticsearch:
		return model.toAPICreateRemoteElasticsearchModel(ctx)
	default:
		return kbapi.NewOutputUnion{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled output type: %s", outputType), ""),
		}
	}
}

func (model outputModel) toAPIUpdateModel(ctx context.Context) (union kbapi.UpdateOutputUnion, diags diag.Diagnostics) {
	outputType := model.Type.ValueString()

	switch outputType {
	case outputTypeElasticsearch:
		return model.toAPIUpdateElasticsearchModel(ctx)
	case outputTypeLogstash:
		return model.toAPIUpdateLogstashModel(ctx)
	case outputTypeKafka:
		return model.toAPIUpdateKafkaModel(ctx)
	case outputTypeRemoteElasticsearch:
		return model.toAPIUpdateRemoteElasticsearchModel(ctx)
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}

func clearRemoteElasticsearchOnlyFields(model *outputModel) {
	model.ServiceToken = types.StringNull()
	model.SyncIntegrations = types.BoolNull()
	model.SyncUninstalledIntegrations = types.BoolNull()
	model.WriteToLogsStreams = types.BoolNull()
}

type commonOutputReadData struct {
	id                   *string
	name                 string
	outputType           string
	hosts                []string
	caSha256             *string
	caTrustedFingerprint *string
	isDefault            *bool
	isDefaultMonitoring  *bool
	configYaml           *string
	ssl                  *kbapi.KibanaHTTPAPIsOutputResponseSsl
}

func (model *outputModel) fromAPICommonFields(ctx context.Context, d commonOutputReadData) (diags diag.Diagnostics) {
	// Capture the existing config_yaml and name before we overwrite the
	// model so we can preserve a user-removed null config_yaml over the
	// Fleet API echo. Fleet treats an omitted config_yaml in update
	// requests as "no change" and echoes the previously stored value (or
	// "" for outputs that never had one) back in the response. Once the
	// user has removed config_yaml from configuration we honour that
	// intent across both update and refresh — otherwise the apply trips
	// "inconsistent values for sensitive attribute" or refresh surfaces
	// perpetual drift (issue #1856). On import the existing model carries
	// only the importer-populated fields (output_id / space_ids), so we
	// use Name being null as the discriminator: a refreshed or
	// post-update model always has a non-null Name from prior state.
	existingConfigYaml := model.ConfigYaml
	isImport := model.Name.IsNull() || model.Name.IsUnknown()

	model.ID = types.StringPointerValue(d.id)
	model.OutputID = types.StringPointerValue(d.id)
	model.Name = types.StringValue(d.name)
	model.Type = types.StringValue(d.outputType)
	model.Hosts = typeutils.SliceToListTypeString(ctx, d.hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(d.caSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(d.caTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(d.isDefault)
	model.DefaultMonitoring = types.BoolPointerValue(d.isDefaultMonitoring)
	model.ConfigYaml = configYamlFromAPI(d.configYaml)
	if !isImport && existingConfigYaml.IsNull() {
		model.ConfigYaml = customtypes.NewNormalizedYamlNull()
	}
	if d.ssl != nil {
		verificationMode := (*kbapi.KibanaHTTPAPIsOutputSslVerificationMode)(nil)
		if d.ssl.VerificationMode != nil {
			mode := kbapi.KibanaHTTPAPIsOutputSslVerificationMode(*d.ssl.VerificationMode)
			verificationMode = &mode
		}
		model.Ssl, diags = sslToObjectValue(ctx, d.ssl.Certificate, d.ssl.CertificateAuthorities, d.ssl.Key, verificationMode)
	} else {
		model.Ssl, diags = sslToObjectValue(ctx, nil, nil, nil, nil)
	}
	if model.SpaceIDs.IsNull() || model.SpaceIDs.IsUnknown() {
		model.SpaceIDs = types.SetNull(types.StringType)
	}
	return
}

type commonNewOutputBody struct {
	CaSha256             *string
	CaTrustedFingerprint *string
	ConfigYaml           *string
	Hosts                []string
	ID                   *string
	IsDefault            *bool
	IsDefaultMonitoring  *bool
	Name                 string
	Ssl                  *kbapi.KibanaHTTPAPIsOutputSsl
}

func (model outputModel) buildCommonNewOutput(ctx context.Context, diags *diag.Diagnostics) commonNewOutputBody {
	ssl, d := objectValueToSSL(ctx, model.Ssl)
	diags.Append(d...)
	return commonNewOutputBody{
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), diags),
		ID:                   typeutils.OptionalString(model.OutputID),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueString(),
		Ssl:                  ssl.toAPI(),
	}
}

type commonUpdateOutputBody struct {
	CaSha256             *string
	CaTrustedFingerprint *string
	ConfigYaml           *string
	Hosts                *[]string
	IsDefault            *bool
	IsDefaultMonitoring  *bool
	Name                 *string
	Ssl                  *kbapi.KibanaHTTPAPIsOutputSsl
}

func (model outputModel) buildCommonUpdateOutput(ctx context.Context, diags *diag.Diagnostics) commonUpdateOutputBody {
	ssl, d := objectValueToSSLUpdate(ctx, model.Ssl)
	diags.Append(d...)
	return commonUpdateOutputBody{
		CaSha256:             model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint: model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:           model.ConfigYaml.ValueStringPointer(),
		Hosts:                typeutils.SliceRef(typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), diags)),
		IsDefault:            model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:  model.DefaultMonitoring.ValueBoolPointer(),
		Name:                 model.Name.ValueStringPointer(),
		Ssl:                  ssl.toAPI(),
	}
}

// configYamlFromAPI normalizes the Fleet API representation of config_yaml
// into the resource's NormalizedYamlValue state attribute. Fleet treats an
// omitted config_yaml as "no change" on update and serializes an unset value
// as an empty string in responses; folding empty strings to null keeps state
// stable across updates that don't touch the field. Issue #1856.
func configYamlFromAPI(value *string) customtypes.NormalizedYamlValue {
	if value == nil || *value == "" {
		return customtypes.NewNormalizedYamlNull()
	}
	return customtypes.NewNormalizedYamlValue(*value)
}

// fromAPISimpleOutput populates a model from the common fields of a simple
// (non-Kafka, non-RemoteElasticsearch) output type and clears the
// remote-Elasticsearch-only fields.
func (model *outputModel) fromAPISimpleOutput(ctx context.Context, d commonOutputReadData) diag.Diagnostics {
	diags := model.fromAPICommonFields(ctx, d)
	clearRemoteElasticsearchOnlyFields(model)
	return diags
}

// toAPICreateSimpleOutput builds a NewOutputUnion for simple output types.
// The caller supplies a buildUnion func that constructs the type-specific body
// and calls the appropriate From* discriminator method.
func (model outputModel) toAPICreateSimpleOutput(
	ctx context.Context,
	buildUnion func(f commonNewOutputBody) (kbapi.NewOutputUnion, error),
) (kbapi.NewOutputUnion, diag.Diagnostics) {
	var diags diag.Diagnostics
	f := model.buildCommonNewOutput(ctx, &diags)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}
	union, err := buildUnion(f)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.NewOutputUnion{}, diags
	}
	return union, diags
}

// toAPIUpdateSimpleOutput builds an UpdateOutputUnion for simple output types.
// The caller supplies a buildUnion func that constructs the type-specific body
// and calls the appropriate From* discriminator method.
func (model outputModel) toAPIUpdateSimpleOutput(
	ctx context.Context,
	buildUnion func(f commonUpdateOutputBody) (kbapi.UpdateOutputUnion, error),
) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	var diags diag.Diagnostics
	f := model.buildCommonUpdateOutput(ctx, &diags)
	if diags.HasError() {
		return kbapi.UpdateOutputUnion{}, diags
	}
	union, err := buildUnion(f)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}
	return union, diags
}
