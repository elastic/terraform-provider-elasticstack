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

package outputds

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputModel struct {
	ID      types.String `tfsdk:"id"`
	SpaceID types.String `tfsdk:"space_id"`
	Outputs types.List   `tfsdk:"outputs"`
}

func (model *outputModel) populateFromApi(ctx context.Context, unions []kbapi.OutputUnion) (diags diag.Diagnostics) {
	model.ID = types.StringValue("outputs")
	model.Outputs = typeutils.SliceToListType(ctx, unions, getOutputItemElemType(), path.Root("outputs"), &diags,
		func(union kbapi.OutputUnion, meta typeutils.ListMeta) outputItemModel {
			model := outputItemModel{}
			diags := model.populateFromAPI(ctx, &union)
			meta.Diags.Append(diags...)
			return model
		})

	return
}

type outputItemModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Hosts                types.List   `tfsdk:"hosts"` // string
	CaSha256             types.String `tfsdk:"ca_sha256"`
	CaTrustedFingerprint types.String `tfsdk:"ca_trusted_fingerprint"`
	DefaultIntegrations  types.Bool   `tfsdk:"default_integrations"`
	DefaultMonitoring    types.Bool   `tfsdk:"default_monitoring"`
	ConfigYaml           types.String `tfsdk:"config_yaml"`
}

func (model *outputItemModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) (diags diag.Diagnostics) {
	if union == nil {
		return
	}

	output, err := union.ValueByDiscriminator()
	if err != nil {
		diags.AddError(err.Error(), "")
		return
	}

	switch output := output.(type) {
	case kbapi.OutputElasticsearch:
		diags.Append(model.fromAPIElasticsearchModel(ctx, &output)...)
	case kbapi.OutputLogstash:
		diags.Append(model.fromAPILogstashModel(ctx, &output)...)
	case kbapi.OutputKafka:
		diags.Append(model.fromAPIKafkaModel(ctx, &output)...)
	case kbapi.OutputRemoteElasticsearch:
		diags.Append(model.fromAPIRemoteElasticsearchModel(ctx, &output)...)
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %T", output), "")
	}

	return
}

func (model *outputItemModel) fromAPIElasticsearchModel(ctx context.Context, data *kbapi.OutputElasticsearch) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Type = types.StringValue(string(data.Type))
	model.Hosts = typeutils.SliceToListTypeString(ctx, data.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(data.CaSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(data.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(data.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(data.IsDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(data.ConfigYaml)
	return
}

func (model *outputItemModel) fromAPIKafkaModel(ctx context.Context, data *kbapi.OutputKafka) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Type = types.StringValue(string(data.Type))
	model.Hosts = typeutils.SliceToListTypeString(ctx, data.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(data.CaSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(data.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(data.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(data.IsDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(data.ConfigYaml)
	return
}

func (model *outputItemModel) fromAPILogstashModel(ctx context.Context, data *kbapi.OutputLogstash) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Type = types.StringValue(string(data.Type))
	model.Hosts = typeutils.SliceToListTypeString(ctx, data.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(data.CaSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(data.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(data.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(data.IsDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(data.ConfigYaml)
	return
}

func (model *outputItemModel) fromAPIRemoteElasticsearchModel(ctx context.Context, data *kbapi.OutputRemoteElasticsearch) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Type = types.StringValue(string(data.Type))
	model.Hosts = typeutils.SliceToListTypeString(ctx, data.Hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(data.CaSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(data.CaTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(data.IsDefault)
	model.DefaultMonitoring = types.BoolPointerValue(data.IsDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(data.ConfigYaml)
	return
}
