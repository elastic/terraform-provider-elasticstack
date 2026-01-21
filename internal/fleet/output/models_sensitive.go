package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// extractSslKeyFromObject extracts the sensitive ssl.key field from an SSL object
func extractSslKeyFromObject(ctx context.Context, obj types.Object) *types.String {
	if !utils.IsKnown(obj) {
		return nil
	}

	var diags diag.Diagnostics
	sslModel := utils.ObjectTypeAs[outputSslModel](ctx, obj, path.Root("ssl"), &diags)
	if diags.HasError() || sslModel == nil {
		return nil
	}

	return &sslModel.Key
}

// restoreSslKeyToObject restores the sensitive ssl.key field to an SSL object
func restoreSslKeyToObject(ctx context.Context, obj types.Object, key types.String, diags *diag.Diagnostics) types.Object {
	if !utils.IsKnown(obj) {
		return obj
	}

	sslModel := utils.ObjectTypeAs[outputSslModel](ctx, obj, path.Root("ssl"), diags)
	if diags.HasError() || sslModel == nil {
		return obj
	}

	// Restore the sensitive key field
	sslModel.Key = key

	// Convert back to object
	result, d := types.ObjectValueFrom(ctx, getSslAttrTypes(), sslModel)
	diags.Append(d...)
	return result
}

// extractKafkaPasswordFromObject extracts the sensitive kafka.password field from a Kafka object
func extractKafkaPasswordFromObject(ctx context.Context, obj types.Object) *types.String {
	if !utils.IsKnown(obj) {
		return nil
	}

	var diags diag.Diagnostics
	kafkaModel := utils.ObjectTypeAs[outputKafkaModel](ctx, obj, path.Root("kafka"), &diags)
	if diags.HasError() || kafkaModel == nil {
		return nil
	}

	return &kafkaModel.Password
}

// restoreKafkaPasswordToObject restores the sensitive kafka.password field to a Kafka object
func restoreKafkaPasswordToObject(ctx context.Context, obj types.Object, password types.String, diags *diag.Diagnostics) types.Object {
	if !utils.IsKnown(obj) {
		return obj
	}

	kafkaModel := utils.ObjectTypeAs[outputKafkaModel](ctx, obj, path.Root("kafka"), diags)
	if diags.HasError() || kafkaModel == nil {
		return obj
	}

	// Restore the sensitive password field
	kafkaModel.Password = password

	// Convert back to object
	result, d := types.ObjectValueFrom(ctx, getKafkaAttrTypes(), kafkaModel)
	diags.Append(d...)
	return result
}
