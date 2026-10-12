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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputModel struct {
	entitycore.KibanaConnectionField
	ID      types.String `tfsdk:"id"`
	SpaceID types.String `tfsdk:"space_id"`
	Outputs types.List   `tfsdk:"outputs"`
}

func (m outputModel) GetID() types.String         { return m.ID }
func (m outputModel) GetResourceID() types.String { return types.StringValue("outputs") }
func (m outputModel) GetSpaceID() types.String    { return m.SpaceID }

func (m *outputModel) populateFromAPI(ctx context.Context, unions []kbapi.OutputUnion) (diags diag.Diagnostics) {
	m.ID = types.StringValue("outputs")
	m.Outputs = typeutils.SliceToListType(ctx, unions, getOutputItemElemType(ctx), path.Root("outputs"), &diags,
		func(union kbapi.OutputUnion, meta typeutils.ListMeta) outputItemModel {
			item := outputItemModel{}
			diags := item.populateFromAPI(ctx, &union)
			meta.Diags.Append(diags...)
			return item
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

func (model *outputItemModel) populateFromAPI(ctx context.Context, union *kbapi.OutputUnion) diag.Diagnostics {
	return fleet.DispatchOutputUnion(ctx, union, fleet.OutputUnionHandlers{
		Elasticsearch:       model.fromAPIElasticsearchModel,
		Logstash:            model.fromAPILogstashModel,
		Kafka:               model.fromAPIKafkaModel,
		RemoteElasticsearch: model.fromAPIRemoteElasticsearchModel,
	})
}

// fromAPICommonFields is the single point of change for the field mapping
// shared across all output API types. The four fromAPI*Model methods below
// exist only to adapt each generated kbapi response type's fields into this
// common call, since Go's type system requires a distinct method per
// concrete parameter type for fleet.OutputUnionHandlers.
func (model *outputItemModel) fromAPICommonFields(ctx context.Context, id *string, name, outputType string, hosts []string, caSha256, caTrustedFingerprint *string, isDefault, isDefaultMonitoring *bool, configYaml *string) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(id)
	model.Name = types.StringValue(name)
	model.Type = types.StringValue(outputType)
	model.Hosts = typeutils.SliceToListTypeString(ctx, hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(caSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(caTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(isDefault)
	model.DefaultMonitoring = types.BoolPointerValue(isDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(configYaml)
	return
}

func (model *outputItemModel) fromAPIElasticsearchModel(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseElasticsearch) diag.Diagnostics {
	return model.fromAPICommonFields(ctx, data.Id, data.Name, string(data.Type), data.Hosts, data.CaSha256, data.CaTrustedFingerprint, data.IsDefault, data.IsDefaultMonitoring, data.ConfigYaml)
}

func (model *outputItemModel) fromAPIKafkaModel(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseKafka) diag.Diagnostics {
	return model.fromAPICommonFields(ctx, data.Id, data.Name, string(data.Type), data.Hosts, data.CaSha256, data.CaTrustedFingerprint, data.IsDefault, data.IsDefaultMonitoring, data.ConfigYaml)
}

func (model *outputItemModel) fromAPILogstashModel(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseLogstash) diag.Diagnostics {
	return model.fromAPICommonFields(ctx, data.Id, data.Name, string(data.Type), data.Hosts, data.CaSha256, data.CaTrustedFingerprint, data.IsDefault, data.IsDefaultMonitoring, data.ConfigYaml)
}

func (model *outputItemModel) fromAPIRemoteElasticsearchModel(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch) diag.Diagnostics {
	return model.fromAPICommonFields(ctx, data.Id, data.Name, string(data.Type), data.Hosts, data.CaSha256, data.CaTrustedFingerprint, data.IsDefault, data.IsDefaultMonitoring, data.ConfigYaml)
}
