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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
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

func (model *outputModel) populateFromAPI(ctx context.Context, unions []kbapi.OutputUnion) (diags diag.Diagnostics) {
	model.ID = types.StringValue("outputs")
	model.Outputs = typeutils.SliceToListType(ctx, unions, getOutputItemElemType(ctx), path.Root("outputs"), &diags,
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
	case kbapi.KibanaHTTPAPIsOutputResponseElasticsearch:
		diags.Append(fromAPIOutputResponse(ctx, model, output.Id, output.Name, output.Type,
			output.Hosts, output.CaSha256, output.CaTrustedFingerprint,
			output.IsDefault, output.IsDefaultMonitoring, output.ConfigYaml)...)
	case kbapi.KibanaHTTPAPIsOutputResponseLogstash:
		diags.Append(fromAPIOutputResponse(ctx, model, output.Id, output.Name, output.Type,
			output.Hosts, output.CaSha256, output.CaTrustedFingerprint,
			output.IsDefault, output.IsDefaultMonitoring, output.ConfigYaml)...)
	case kbapi.KibanaHTTPAPIsOutputResponseKafka:
		diags.Append(fromAPIOutputResponse(ctx, model, output.Id, output.Name, output.Type,
			output.Hosts, output.CaSha256, output.CaTrustedFingerprint,
			output.IsDefault, output.IsDefaultMonitoring, output.ConfigYaml)...)
	case kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch:
		diags.Append(fromAPIOutputResponse(ctx, model, output.Id, output.Name, output.Type,
			output.Hosts, output.CaSha256, output.CaTrustedFingerprint,
			output.IsDefault, output.IsDefaultMonitoring, output.ConfigYaml)...)
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %T", output), "")
	}

	return
}

// outputAPICommonData holds the fields shared across all output API types so
// that fromAPICommonFields can serve as the single point of change for the
// common field mapping logic.
type outputAPICommonData struct {
	id                   *string
	name                 string
	outputType           string
	hosts                []string
	caSha256             *string
	caTrustedFingerprint *string
	isDefault            *bool
	isDefaultMonitoring  *bool
	configYaml           *string
}

func (model *outputItemModel) fromAPICommonFields(ctx context.Context, d outputAPICommonData) (diags diag.Diagnostics) {
	model.ID = types.StringPointerValue(d.id)
	model.Name = types.StringValue(d.name)
	model.Type = types.StringValue(d.outputType)
	model.Hosts = typeutils.SliceToListTypeString(ctx, d.hosts, path.Root("hosts"), &diags)
	model.CaSha256 = types.StringPointerValue(d.caSha256)
	model.CaTrustedFingerprint = typeutils.NonEmptyStringishPointerValue(d.caTrustedFingerprint)
	model.DefaultIntegrations = types.BoolPointerValue(d.isDefault)
	model.DefaultMonitoring = types.BoolPointerValue(d.isDefaultMonitoring)
	model.ConfigYaml = types.StringPointerValue(d.configYaml)
	return
}

// outputResponseType is the set of generated Fleet output "type" enums
// (KibanaHTTPAPIsOutputResponse{Elasticsearch,Kafka,Logstash,RemoteElasticsearch}Type).
// They are all distinct named types so Go can't unify field access across the
// response structs directly, but every one of them is string-based, which lets
// fromAPIOutputResponse stay generic over just that one varying piece.
type outputResponseType interface {
	~string
}

// fromAPIOutputResponse maps the fields shared by every kbapi output response
// type into outputAPICommonData. The four KibanaHTTPAPIsOutputResponse* structs
// are separately generated and not structurally identical (each has its own
// extra fields), so Go generics can't select their fields directly; callers
// extract the shared fields themselves and this is the single point of change
// for turning them into outputAPICommonData. Go methods can't declare their own
// type parameters, so this is a plain function taking model explicitly.
func fromAPIOutputResponse[T outputResponseType](
	ctx context.Context, model *outputItemModel,
	id *string, name string, outputType T, hosts []string,
	caSha256, caTrustedFingerprint *string,
	isDefault, isDefaultMonitoring *bool,
	configYaml *string,
) (diags diag.Diagnostics) {
	return model.fromAPICommonFields(ctx, outputAPICommonData{
		id: id, name: name, outputType: string(outputType),
		hosts: hosts, caSha256: caSha256,
		caTrustedFingerprint: caTrustedFingerprint,
		isDefault:            isDefault, isDefaultMonitoring: isDefaultMonitoring,
		configYaml: configYaml,
	})
}
