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
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
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
	return fleet.SpaceIDFromSetOrDefault(model.SpaceIDs, "")
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

func (model *outputModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) diag.Diagnostics {
	return fleet.DispatchOutputUnion(ctx, union, fleet.OutputUnionHandlers{
		Elasticsearch:       model.fromAPIElasticsearchModel,
		Logstash:            model.fromAPILogstashModel,
		Kafka:               model.fromAPIKafkaModel,
		RemoteElasticsearch: model.fromAPIRemoteElasticsearchModel,
	})
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
	ID                   *string
	Name                 string
	OutputType           string
	Hosts                []string
	CaSha256             *string
	CaTrustedFingerprint *string
	IsDefault            *bool
	IsDefaultMonitoring  *bool
	ConfigYaml           *string
	Ssl                  *kbapi.KibanaHTTPAPIsOutputResponseSsl
}

// outputReadFieldNameOverrides maps a commonOutputReadData field name to the
// differently-named field on the generated kbapi output response structs.
var outputReadFieldNameOverrides = map[string]string{
	"ID":         "Id",
	"OutputType": "Type",
}

// buildCommonOutputReadData reflects over data (a pointer to a generated
// kbapi output response struct, e.g. KibanaHTTPAPIsOutputResponseElasticsearch
// or KibanaHTTPAPIsOutputResponseLogstash) and copies every commonOutputReadData
// field from the identically-named (or overridden) field on data, converting
// between the generated per-type Type enum and the plain string OutputType.
// This replicates what every simple output type's fromAPI builder previously
// did by hand in a ~10-field struct literal.
//
// A field present on both structs under a non-convertible type is a
// programmer error (not a runtime condition), so it panics rather than
// silently dropping the field.
func buildCommonOutputReadData[T any](data *T) commonOutputReadData {
	var d commonOutputReadData
	dVal := reflect.ValueOf(&d).Elem()
	dType := dVal.Type()
	dataVal := reflect.ValueOf(data).Elem()

	for i := range dType.NumField() {
		field := dType.Field(i)
		srcName := field.Name
		if override, ok := outputReadFieldNameOverrides[field.Name]; ok {
			srcName = override
		}

		source := dataVal.FieldByName(srcName)
		if !source.IsValid() {
			continue
		}

		if !source.Type().ConvertibleTo(field.Type) {
			panic(fmt.Sprintf("output: common field %q type mismatch: want %s, got %s on %T", field.Name, field.Type, source.Type(), data))
		}

		dVal.Field(i).Set(source.Convert(field.Type))
	}

	return d
}

// populateCommonOutputFields reflects over target (a pointer to a generated
// kbapi output New/Update struct, e.g. KibanaHTTPAPIsNewOutputElasticsearch or
// KibanaHTTPAPIsUpdateOutputLogstash) and sets every field on target whose
// name (ID on src maps to Id on target) and type match a field on src (a
// commonNewOutputBody or commonUpdateOutputBody), replicating what every
// simple output type's toAPICreate/toAPIUpdate builder previously did by hand
// in a ~8-field struct literal. Fields on target that have no counterpart on
// src (e.g. the per-type Type discriminator, or type-specific fields) are
// left untouched for the caller to set directly.
//
// A field present on both structs under a mismatched type is a programmer
// error (not a runtime condition), so it panics rather than silently dropping
// the field.
func populateCommonOutputFields[T any](target *T, src any) {
	targetVal := reflect.ValueOf(target).Elem()
	srcVal := reflect.ValueOf(src)
	srcType := srcVal.Type()

	for i := range srcType.NumField() {
		field := srcType.Field(i)
		name := field.Name
		if name == "ID" {
			name = "Id"
		}

		targetField := targetVal.FieldByName(name)
		if !targetField.IsValid() {
			continue
		}

		if targetField.Type() != field.Type {
			panic(fmt.Sprintf("output: common field %q type mismatch: want %s, got %s on %T", name, targetField.Type(), field.Type, target))
		}

		targetField.Set(srcVal.Field(i))
	}
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

	model.ID = types.StringPointerValue(d.ID)
	model.OutputID = types.StringPointerValue(d.ID)
	model.Name = types.StringValue(d.Name)
	model.Type = types.StringValue(d.OutputType)
	model.Hosts = typeutils.SliceToListTypeString(ctx, d.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(d.CaSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(d.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(d.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(d.IsDefaultMonitoring)
	model.ConfigYaml = configYamlFromAPI(d.ConfigYaml)
	if !isImport && existingConfigYaml.IsNull() {
		model.ConfigYaml = customtypes.NewNormalizedYamlNull()
	}
	if d.Ssl != nil {
		verificationMode := (*kbapi.KibanaHTTPAPIsOutputSslVerificationMode)(nil)
		if d.Ssl.VerificationMode != nil {
			mode := kbapi.KibanaHTTPAPIsOutputSslVerificationMode(*d.Ssl.VerificationMode)
			verificationMode = &mode
		}
		model.Ssl, diags = sslToObjectValue(ctx, d.Ssl.Certificate, d.Ssl.CertificateAuthorities, d.Ssl.Key, verificationMode)
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
