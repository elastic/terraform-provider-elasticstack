package output

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractSslKeyFromObject(t *testing.T) {
	ctx := context.Background()

	t.Run("returns nil when object is null", func(t *testing.T) {
		obj := types.ObjectNull(getSslAttrTypes())
		result := extractSslKeyFromObject(ctx, obj)
		assert.Nil(t, result)
	})

	t.Run("returns nil when object is unknown", func(t *testing.T) {
		obj := types.ObjectUnknown(getSslAttrTypes())
		result := extractSslKeyFromObject(ctx, obj)
		assert.Nil(t, result)
	})

	t.Run("extracts key field from ssl object", func(t *testing.T) {
		sslModel := outputSslModel{
			Certificate:            types.StringValue("cert-content"),
			Key:                    types.StringValue("key-content"),
			CertificateAuthorities: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca1")}),
		}
		obj, diags := types.ObjectValueFrom(ctx, getSslAttrTypes(), sslModel)
		require.False(t, diags.HasError())

		result := extractSslKeyFromObject(ctx, obj)
		require.NotNil(t, result)
		assert.Equal(t, "key-content", result.ValueString())
	})
}

func TestRestoreSslKeyToObject(t *testing.T) {
	ctx := context.Background()

	t.Run("returns same object when null", func(t *testing.T) {
		obj := types.ObjectNull(getSslAttrTypes())
		key := types.StringValue("new-key")
		var diags diag.Diagnostics

		result := restoreSslKeyToObject(ctx, obj, key, &diags)
		assert.False(t, diags.HasError())
		assert.True(t, result.IsNull())
	})

	t.Run("restores key field in ssl object", func(t *testing.T) {
		// Simulate API response which nulls the key
		sslModelFromAPI := outputSslModel{
			Certificate:            types.StringValue("cert-content"),
			Key:                    types.StringNull(),
			CertificateAuthorities: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ca1")}),
		}
		objFromAPI, diags := types.ObjectValueFrom(ctx, getSslAttrTypes(), sslModelFromAPI)
		require.False(t, diags.HasError())

		// Restore the key
		restoredKey := types.StringValue("original-key")
		result := restoreSslKeyToObject(ctx, objFromAPI, restoredKey, &diags)
		assert.False(t, diags.HasError())

		// Verify the key was restored
		var resultModel outputSslModel
		diags = result.As(ctx, &resultModel, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError())
		assert.Equal(t, "original-key", resultModel.Key.ValueString())
		assert.Equal(t, "cert-content", resultModel.Certificate.ValueString())
	})
}

func TestExtractKafkaPasswordFromObject(t *testing.T) {
	ctx := context.Background()

	t.Run("returns nil when object is null", func(t *testing.T) {
		obj := types.ObjectNull(getKafkaAttrTypes())
		result := extractKafkaPasswordFromObject(ctx, obj)
		assert.Nil(t, result)
	})

	t.Run("returns nil when object is unknown", func(t *testing.T) {
		obj := types.ObjectUnknown(getKafkaAttrTypes())
		result := extractKafkaPasswordFromObject(ctx, obj)
		assert.Nil(t, result)
	})

	t.Run("extracts password field from kafka object", func(t *testing.T) {
		kafkaModel := outputKafkaModel{
			AuthType:       types.StringValue("user_pass"),
			Username:       types.StringValue("user"),
			Password:       types.StringValue("secret-password"),
			Topic:          types.StringValue("test-topic"),
			BrokerTimeout:  types.Float32Null(),
			ClientId:       types.StringNull(),
			Compression:    types.StringNull(),
			CompressionLevel: types.Int64Null(),
			ConnectionType: types.StringNull(),
			Partition:      types.StringNull(),
			RequiredAcks:   types.Int64Null(),
			Timeout:        types.Float32Null(),
			Version:        types.StringNull(),
			Key:            types.StringNull(),
			Headers:        types.ListNull(getHeadersAttrTypes()),
			Hash:           types.ObjectNull(getHashAttrTypes()),
			Random:         types.ObjectNull(getRandomAttrTypes()),
			RoundRobin:     types.ObjectNull(getRoundRobinAttrTypes()),
			Sasl:           types.ObjectNull(getSaslAttrTypes()),
		}
		obj, diags := types.ObjectValueFrom(ctx, getKafkaAttrTypes(), kafkaModel)
		require.False(t, diags.HasError(), "Failed to create kafka object: %v", diags.Errors())

		result := extractKafkaPasswordFromObject(ctx, obj)
		require.NotNil(t, result)
		assert.Equal(t, "secret-password", result.ValueString())
	})
}

func TestRestoreKafkaPasswordToObject(t *testing.T) {
	ctx := context.Background()

	t.Run("returns same object when null", func(t *testing.T) {
		obj := types.ObjectNull(getKafkaAttrTypes())
		password := types.StringValue("new-password")
		var diags diag.Diagnostics

		result := restoreKafkaPasswordToObject(ctx, obj, password, &diags)
		assert.False(t, diags.HasError())
		assert.True(t, result.IsNull())
	})

	t.Run("restores password field in kafka object", func(t *testing.T) {
		// Simulate API response which nulls the password
		kafkaModelFromAPI := outputKafkaModel{
			AuthType:       types.StringValue("user_pass"),
			Username:       types.StringValue("user"),
			Password:       types.StringNull(),
			Topic:          types.StringValue("test-topic"),
			BrokerTimeout:  types.Float32Null(),
			ClientId:       types.StringNull(),
			Compression:    types.StringNull(),
			CompressionLevel: types.Int64Null(),
			ConnectionType: types.StringNull(),
			Partition:      types.StringNull(),
			RequiredAcks:   types.Int64Null(),
			Timeout:        types.Float32Null(),
			Version:        types.StringNull(),
			Key:            types.StringNull(),
			Headers:        types.ListNull(getHeadersAttrTypes()),
			Hash:           types.ObjectNull(getHashAttrTypes()),
			Random:         types.ObjectNull(getRandomAttrTypes()),
			RoundRobin:     types.ObjectNull(getRoundRobinAttrTypes()),
			Sasl:           types.ObjectNull(getSaslAttrTypes()),
		}
		objFromAPI, diags := types.ObjectValueFrom(ctx, getKafkaAttrTypes(), kafkaModelFromAPI)
		require.False(t, diags.HasError(), "Failed to create kafka object: %v", diags.Errors())

		// Restore the password
		restoredPassword := types.StringValue("original-password")
		var restoreDiags diag.Diagnostics
		result := restoreKafkaPasswordToObject(ctx, objFromAPI, restoredPassword, &restoreDiags)
		assert.False(t, restoreDiags.HasError(), "Failed to restore password: %v", restoreDiags.Errors())

		// Verify the password was restored but other fields remain unchanged
		var resultModel outputKafkaModel
		diags = result.As(ctx, &resultModel, basetypes.ObjectAsOptions{})
		require.False(t, diags.HasError(), "Failed to convert result: %v", diags.Errors())
		assert.Equal(t, "original-password", resultModel.Password.ValueString())
		assert.Equal(t, "user", resultModel.Username.ValueString())
		assert.Equal(t, "test-topic", resultModel.Topic.ValueString())
	})
}
