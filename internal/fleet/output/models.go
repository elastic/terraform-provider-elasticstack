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
	ID                   types.String `tfsdk:"id"`
	OutputID             types.String `tfsdk:"output_id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Hosts                types.List   `tfsdk:"hosts"` //> string
	CaSha256             types.String `tfsdk:"ca_sha256"`
	CaTrustedFingerprint types.String `tfsdk:"ca_trusted_fingerprint"`
	DefaultIntegrations  types.Bool   `tfsdk:"default_integrations"`
	DefaultMonitoring    types.Bool   `tfsdk:"default_monitoring"`
	ConfigYaml           types.String `tfsdk:"config_yaml"`
	Ssl                  types.Object `tfsdk:"ssl"`   //> outputSslModel
	Kafka                types.Object `tfsdk:"kafka"` //> outputKafkaModel
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
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %T", output), "")
	}

	return
}

func (model outputModel) toAPICreateModel(ctx context.Context, client *clients.ApiClient) (kbapi.NewOutputUnion, diag.Diagnostics) {
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
	default:
		return kbapi.NewOutputUnion{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled output type: %s", outputType), ""),
		}
	}
}

func (model outputModel) toAPIUpdateModel(ctx context.Context, client *clients.ApiClient) (union kbapi.UpdateOutputUnion, diags diag.Diagnostics) {
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
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %s", outputType), "")
	}

	return
}

func assertKafkaSupport(ctx context.Context, client *clients.ApiClient) diag.Diagnostics {
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
