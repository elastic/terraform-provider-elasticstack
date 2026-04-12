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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *outputModel) fromAPIRemoteElasticsearchModel(ctx context.Context, data *kbapi.OutputRemoteElasticsearch) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(data.Id)
	model.OutputID = types.StringPointerValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Type = types.StringValue(string(data.Type))
	model.Hosts = typeutils.SliceToListTypeString(ctx, data.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(data.CaSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(data.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(data.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(data.IsDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(data.ConfigYaml)
	if data.Ssl != nil {
		model.Ssl, diags = sslToObjectValue(ctx, data.Ssl.Certificate, data.Ssl.CertificateAuthorities, data.Ssl.Key)
	} else {
		model.Ssl, diags = sslToObjectValue(ctx, nil, nil, nil)
	}

	// Preserve configured secret when Fleet omits/redacts it in read responses.
	if data.ServiceToken != nil {
		model.ServiceToken = types.StringPointerValue(data.ServiceToken)
	} else if model.ServiceToken.IsNull() || model.ServiceToken.IsUnknown() {
		model.ServiceToken = types.StringNull()
	}

	model.SyncIntegrations = types.BoolPointerValue(data.SyncIntegrations)
	model.SyncUninstalledIntegrations = types.BoolPointerValue(data.SyncUninstalledIntegrations)
	model.WriteToLogsStreams = types.BoolPointerValue(data.WriteToLogsStreams)

	// Note: SpaceIDs is not returned by the API for outputs.
	if model.SpaceIDs.IsNull() || model.SpaceIDs.IsUnknown() {
		model.SpaceIDs = types.SetNull(types.StringType)
	}

	return
}

func (model outputModel) toAPICreateRemoteElasticsearchModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSL(ctx, model.Ssl)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}

	body := kbapi.NewOutputRemoteElasticsearch{
		Type:                        kbapi.KibanaHTTPAPIsNewOutputRemoteElasticsearchTypeRemoteElasticsearch,
		CaSha256:                    model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint:        model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:                  model.ConfigYaml.ValueStringPointer(),
		Hosts:                       typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), &diags),
		Id:                          model.OutputID.ValueStringPointer(),
		IsDefault:                   model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:         model.DefaultMonitoring.ValueBoolPointer(),
		Name:                        model.Name.ValueString(),
		ServiceToken:                model.ServiceToken.ValueStringPointer(),
		Ssl:                         ssl.toCreateRemoteElasticsearch(),
		SyncIntegrations:            model.SyncIntegrations.ValueBoolPointer(),
		SyncUninstalledIntegrations: model.SyncUninstalledIntegrations.ValueBoolPointer(),
		WriteToLogsStreams:          model.WriteToLogsStreams.ValueBoolPointer(),
	}

	var union kbapi.NewOutputUnion
	err := union.FromNewOutputRemoteElasticsearch(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.NewOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPIUpdateRemoteElasticsearchModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	ssl, diags := objectValueToSSLUpdate(ctx, model.Ssl)
	if diags.HasError() {
		return kbapi.UpdateOutputUnion{}, diags
	}

	body := kbapi.UpdateOutputRemoteElasticsearch{
		Type: func() *kbapi.KibanaHTTPAPIsUpdateOutputRemoteElasticsearchType {
			outputType := kbapi.RemoteElasticsearch
			return &outputType
		}(),
		CaSha256:                    model.CaSha256.ValueStringPointer(),
		CaTrustedFingerprint:        model.CaTrustedFingerprint.ValueStringPointer(),
		ConfigYaml:                  model.ConfigYaml.ValueStringPointer(),
		Hosts:                       schemautil.SliceRef(typeutils.ListTypeToSliceString(ctx, model.Hosts, path.Root("hosts"), &diags)),
		IsDefault:                   model.DefaultIntegrations.ValueBoolPointer(),
		IsDefaultMonitoring:         model.DefaultMonitoring.ValueBoolPointer(),
		Name:                        model.Name.ValueStringPointer(),
		ServiceToken:                model.ServiceToken.ValueStringPointer(),
		Ssl:                         ssl.toUpdateRemoteElasticsearch(),
		SyncIntegrations:            model.SyncIntegrations.ValueBoolPointer(),
		SyncUninstalledIntegrations: model.SyncUninstalledIntegrations.ValueBoolPointer(),
		WriteToLogsStreams:          model.WriteToLogsStreams.ValueBoolPointer(),
	}

	var union kbapi.UpdateOutputUnion
	err := union.FromUpdateOutputRemoteElasticsearch(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}

	return union, diags
}
