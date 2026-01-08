package customtypes

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONWithContextType_String(t *testing.T) {
	typ := NewJSONWithContextualDefaultsType(nil)
	assert.Equal(t, "customtypes.JSONWithContextType", typ.String())
}

func TestJSONWithContextType_ValueType(t *testing.T) {
	typ := NewJSONWithContextualDefaultsType(nil)
	val := typ.ValueType(context.Background())
	assert.IsType(t, JSONWithContextualDefaultsValue{}, val)
}

func TestJSONWithContextType_Equal(t *testing.T) {
	typ1 := NewJSONWithContextualDefaultsType(nil)
	typ2 := NewJSONWithContextualDefaultsType(nil)

	assert.True(t, typ1.Equal(typ2))
	assert.False(t, typ1.Equal(basetypes.StringType{}))
}

func TestJSONWithContextType_ValueFromString(t *testing.T) {
	ctx := context.Background()
	typ := NewJSONWithContextualDefaultsType(nil)

	t.Run("Valid JSON with context key", func(t *testing.T) {
		jsonStr := `{"key": "value", "__tf_provider_context": "ctx1"}`
		val, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(jsonStr))
		require.False(t, diags.HasError())
		assert.Equal(t, "ctx1", val.(JSONWithContextualDefaultsValue).contextValue)
	})

	t.Run("Valid JSON with legacy context key", func(t *testing.T) {
		jsonStr := `{"key": "value", "__tf_provider_connector_type_id": "ctx2"}`
		val, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(jsonStr))
		require.False(t, diags.HasError())
		assert.Equal(t, "ctx2", val.(JSONWithContextualDefaultsValue).contextValue)
	})

	t.Run("Valid JSON without context key", func(t *testing.T) {
		jsonStr := `{"key": "value"}`
		val, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(jsonStr))
		require.False(t, diags.HasError())
		assert.Empty(t, val.(JSONWithContextualDefaultsValue).contextValue)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		jsonStr := `invalid-json`
		_, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(jsonStr))
		assert.True(t, diags.HasError())
	})

	t.Run("Unknown value", func(t *testing.T) {
		val, diags := typ.ValueFromString(ctx, basetypes.NewStringUnknown())
		require.False(t, diags.HasError())
		assert.True(t, val.IsUnknown())
	})

	t.Run("Null value", func(t *testing.T) {
		val, diags := typ.ValueFromString(ctx, basetypes.NewStringNull())
		require.False(t, diags.HasError())
		assert.True(t, val.IsNull())
	})
}

func TestJSONWithContextType_ValueFromTerraform(t *testing.T) {
	ctx := context.Background()
	typ := NewJSONWithContextualDefaultsType(nil)

	t.Run("Valid string value", func(t *testing.T) {
		jsonStr := `{"key": "value"}`
		val, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(tftypes.String, jsonStr))
		require.NoError(t, err)
		assert.Equal(t, jsonStr, val.(JSONWithContextualDefaultsValue).ValueString())
	})

	t.Run("Invalid type", func(t *testing.T) {
		_, err := typ.ValueFromTerraform(ctx, tftypes.NewValue(tftypes.Number, 123))
		assert.Error(t, err)
	})
}

func TestJSONWithContext_Type(t *testing.T) {
	val := NewJSONWithContextualDefaultsNull()
	assert.IsType(t, JSONWithContextualDefaultsType{}, val.Type(context.Background()))
}

func TestJSONWithContext_Equal(t *testing.T) {
	val1, _ := NewJSONWithContextualDefaultsValue(`{"a":1}`, "ctx", nil)
	val2, _ := NewJSONWithContextualDefaultsValue(`{"a":1}`, "ctx", nil)
	val3, _ := NewJSONWithContextualDefaultsValue(`{"a":2}`, "ctx", nil)

	assert.True(t, val1.Equal(val2))
	assert.False(t, val1.Equal(val3))
	assert.False(t, val1.Equal(basetypes.NewStringValue("")))
}

func TestJSONWithContext_SanitizedValue(t *testing.T) {
	t.Run("Removes context keys and nulls", func(t *testing.T) {
		jsonStr := `{"key": "value", "nullKey": null, "__tf_provider_context": "ctx", "__tf_provider_connector_type_id": "legacy"}`
		val, _ := NewJSONWithContextualDefaultsValue(jsonStr, "ctx", nil)

		sanitized, diags := val.SanitizedValue()
		require.False(t, diags.HasError())

		var m map[string]interface{}
		err := json.Unmarshal([]byte(sanitized), &m)
		require.NoError(t, err)

		assert.Equal(t, "value", m["key"])
		assert.NotContains(t, m, "nullKey")
		assert.NotContains(t, m, "__tf_provider_context")
		assert.NotContains(t, m, "__tf_provider_connector_type_id")
	})

	t.Run("Null value", func(t *testing.T) {
		val := NewJSONWithContextualDefaultsNull()
		sanitized, diags := val.SanitizedValue()
		require.False(t, diags.HasError())
		assert.Empty(t, sanitized)
	})

	t.Run("Unknown value", func(t *testing.T) {
		val := NewJSONWithContextualDefaultsUnknown()
		sanitized, diags := val.SanitizedValue()
		require.False(t, diags.HasError())
		assert.Empty(t, sanitized)
	})
}

func TestJSONWithContext_StringSemanticEquals(t *testing.T) {
	ctx := context.Background()
	populateDefaults := func(contextValue string, value string) (string, error) {
		if contextValue == "error" {
			return "", assert.AnError
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(value), &m); err != nil {
			return "", err
		}
		if m["default"] == nil {
			m["default"] = "value"
		}
		b, err := json.Marshal(m)
		return string(b), err
	}

	t.Run("Equal with defaults", func(t *testing.T) {
		val1, _ := NewJSONWithContextualDefaultsValue(`{"a": 1}`, "ctx", populateDefaults)
		val2, _ := NewJSONWithContextualDefaultsValue(`{"a": 1, "default": "value"}`, "ctx", populateDefaults)

		equal, diags := val1.StringSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.True(t, equal)
	})

	t.Run("Not equal with defaults", func(t *testing.T) {
		val1, _ := NewJSONWithContextualDefaultsValue(`{"a": 1}`, "ctx", populateDefaults)
		val2, _ := NewJSONWithContextualDefaultsValue(`{"a": 1, "default": "other"}`, "ctx", populateDefaults)

		equal, diags := val1.StringSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.False(t, equal)
	})

	t.Run("Fallback without populateDefaults", func(t *testing.T) {
		val1, _ := NewJSONWithContextualDefaultsValue(`{"a": 1}`, "ctx", nil)
		val2, _ := NewJSONWithContextualDefaultsValue(`{"a": 1}`, "ctx", nil)

		equal, diags := val1.StringSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.True(t, equal)
	})

	t.Run("Error in populateDefaults", func(t *testing.T) {
		val1, _ := NewJSONWithContextualDefaultsValue(`{"a": 1}`, "error", populateDefaults)
		val2, _ := NewJSONWithContextualDefaultsValue(`{"a": 1}`, "error", populateDefaults)

		_, diags := val1.StringSemanticEquals(ctx, val2)
		assert.True(t, diags.HasError())
	})

	t.Run("Null and Unknown", func(t *testing.T) {
		valNull := NewJSONWithContextualDefaultsNull()
		valUnknown := NewJSONWithContextualDefaultsUnknown()
		valKnown, _ := NewJSONWithContextualDefaultsValue(`{}`, "ctx", nil)

		eq, _ := valNull.StringSemanticEquals(ctx, NewJSONWithContextualDefaultsNull())
		assert.True(t, eq)

		eq, _ = valNull.StringSemanticEquals(ctx, valKnown)
		assert.False(t, eq)

		eq, _ = valUnknown.StringSemanticEquals(ctx, NewJSONWithContextualDefaultsUnknown())
		assert.True(t, eq)
	})
}

func TestNewJSONWithContext(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		val, diags := NewJSONWithContextualDefaultsValue(`{"key": "value"}`, "ctx", nil)
		require.False(t, diags.HasError())
		assert.False(t, val.IsNull())
		assert.False(t, val.IsUnknown())
		assert.Equal(t, "ctx", val.contextValue)

		// Check if context key was added to the string value
		var m map[string]interface{}
		err := json.Unmarshal([]byte(val.ValueString()), &m)
		require.NoError(t, err)
		assert.Equal(t, "ctx", m["__tf_provider_context"])
	})

	t.Run("Empty value returns Null", func(t *testing.T) {
		val, diags := NewJSONWithContextualDefaultsValue("", "ctx", nil)
		require.False(t, diags.HasError())
		assert.True(t, val.IsNull())
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		_, diags := NewJSONWithContextualDefaultsValue(`invalid`, "ctx", nil)
		assert.True(t, diags.HasError())
	})
}
