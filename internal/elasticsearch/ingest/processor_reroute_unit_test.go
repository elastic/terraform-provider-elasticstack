package ingest

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceProcessorReroute_Unit(t *testing.T) {
	// Test that we can create and call the data source
	resource := DataSourceProcessorReroute()

	assert.NotNil(t, resource)
	assert.Contains(t, resource.Description, "reroute")
	assert.Contains(t, resource.Schema, "field")
	assert.Contains(t, resource.Schema, "ignore_missing")
	assert.Contains(t, resource.Schema, "json")

	// Test data source read function
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"field":          "routing_field",
		"ignore_missing": true,
		"description":    "Test reroute processor",
	})

	ctx := context.Background()
	diags := resource.ReadContext(ctx, d, nil)

	assert.False(t, diags.HasError(), "Data source read should not have errors")
	assert.NotEmpty(t, d.Get("json"))
	assert.NotEmpty(t, d.Id())

	jsonOutput := d.Get("json").(string)
	assert.Contains(t, jsonOutput, "reroute")
	assert.Contains(t, jsonOutput, "routing_field")
	assert.Contains(t, jsonOutput, "ignore_missing")
}

func TestDataSourceProcessorReroute_MinimalConfig(t *testing.T) {
	resource := DataSourceProcessorReroute()

	// Test with just the required field
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"field": "minimal_field",
	})

	ctx := context.Background()
	diags := resource.ReadContext(ctx, d, nil)

	assert.False(t, diags.HasError(), "Data source read should not have errors")
	assert.NotEmpty(t, d.Get("json"))

	jsonOutput := d.Get("json").(string)
	assert.Contains(t, jsonOutput, "minimal_field")
	assert.Contains(t, jsonOutput, "ignore_failure")
	assert.Contains(t, jsonOutput, "ignore_missing")
}

func TestDataSourceProcessorReroute_AllFields(t *testing.T) {
	resource := DataSourceProcessorReroute()

	// Test with all optional fields
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"field":          "all_fields_test",
		"ignore_missing": true,
		"description":    "Full processor test",
		"if":             "ctx.field != null",
		"ignore_failure": true,
		"tag":            "reroute-tag",
		"on_failure":     []interface{}{`{"set": {"field": "error", "value": "reroute_failed"}}`},
	})

	ctx := context.Background()
	diags := resource.ReadContext(ctx, d, nil)

	assert.False(t, diags.HasError(), "Data source read should not have errors")
	assert.NotEmpty(t, d.Get("json"))

	jsonOutput := d.Get("json").(string)
	assert.Contains(t, jsonOutput, "all_fields_test")
	assert.Contains(t, jsonOutput, "Full processor test")
	assert.Contains(t, jsonOutput, "ctx.field != null")
	assert.Contains(t, jsonOutput, "reroute-tag")
	assert.Contains(t, jsonOutput, "on_failure")
}