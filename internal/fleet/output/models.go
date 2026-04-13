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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outputModel struct {
	ID                          types.String `tfsdk:"id"`
	OutputID                    types.String `tfsdk:"output_id"`
	Name                        types.String `tfsdk:"name"`
	Type                        types.String `tfsdk:"type"`
	Hosts                       types.List   `tfsdk:"hosts"` // > string
	ServiceToken                types.String `tfsdk:"service_token"`
	CaSha256                    types.String `tfsdk:"ca_sha256"`
	CaTrustedFingerprint        types.String `tfsdk:"ca_trusted_fingerprint"`
	DefaultIntegrations         types.Bool   `tfsdk:"default_integrations"`
	DefaultMonitoring           types.Bool   `tfsdk:"default_monitoring"`
	ConfigYaml                  types.String `tfsdk:"config_yaml"`
	SpaceIDs                    types.Set    `tfsdk:"space_ids"` // > string
	Ssl                         types.Object `tfsdk:"ssl"`       // > outputSslModel
	Kafka                       types.Object `tfsdk:"kafka"`     // > outputKafkaModel
	SyncIntegrations            types.Bool   `tfsdk:"sync_integrations"`
	SyncUninstalledIntegrations types.Bool   `tfsdk:"sync_uninstalled_integrations"`
	WriteToLogsStreams          types.Bool   `tfsdk:"write_to_logs_streams"`
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

func (model outputModel) toAPICreateModel(ctx context.Context, client *clients.APIClient) (kbapi.NewOutputUnion, diag.Diagnostics) {
	outputType := model.Type.ValueString()

	switch outputType {
	case "elasticsearch":
		return model.toAPICreateElasticsearchModel(ctx)
	case "logstash":
		return model.toAPICreateLogstashModel(ctx)
	case "kafka":
		if diags := assertKafkaSupport(ctx, client); diags.HasError() {
			return kbapi.NewOutputUnion{}, diags
		}

		return model.toAPICreateKafkaModel(ctx)
	case "remote_elasticsearch":
		return model.toAPICreateRemoteElasticsearchModel(ctx)
	default:
		return kbapi.NewOutputUnion{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled output type: %s", outputType), ""),
		}
	}
}

func (model outputModel) toAPIUpdateModel(ctx context.Context, client *clients.APIClient) (union kbapi.UpdateOutputUnion, diags diag.Diagnostics) {
	outputType := model.Type.ValueString()

	switch outputType {
	case "elasticsearch":
		return model.toAPIUpdateElasticsearchModel(ctx)
	case "logstash":
		return model.toAPIUpdateLogstashModel(ctx)
	case "kafka":
		if diags := assertKafkaSupport(ctx, client); diags.HasError() {
			return kbapi.UpdateOutputUnion{}, diags
		}

		return model.toAPIUpdateKafkaModel(ctx)
	case "remote_elasticsearch":
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

func assertKafkaSupport(ctx context.Context, client *clients.APIClient) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check minimum version requirement for Kafka output type
	if supported, versionDiags := client.EnforceMinVersion(ctx, MinVersionOutputKafka); versionDiags.HasError() {
		diags.Append(diagutil.FrameworkDiagsFromSDK(versionDiags)...)
		return diags
	} else if !supported {
		diags.AddError("Unsupported version for Kafka output",
			fmt.Sprintf("Kafka output type requires server version %s or higher", MinVersionOutputKafka.String()))
		return diags
	}

	return nil
}
