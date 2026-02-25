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
	assert.Contains(t, resource.Schema, "destination")
	assert.Contains(t, resource.Schema, "dataset")
	assert.Contains(t, resource.Schema, "namespace")
	assert.Contains(t, resource.Schema, "json")

	// Test data source read function
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"destination": "target-index",
		"dataset":     "logs",
		"namespace":   "production",
		"description": "Test reroute processor",
	})

	ctx := context.Background()
	diags := resource.ReadContext(ctx, d, nil)

	assert.False(t, diags.HasError(), "Data source read should not have errors")
	assert.NotEmpty(t, d.Get("json"))
	assert.NotEmpty(t, d.Id())

	jsonOutput := d.Get("json").(string)
	assert.Contains(t, jsonOutput, "reroute")
	assert.Contains(t, jsonOutput, "target-index")
	assert.Contains(t, jsonOutput, "logs")
	assert.Contains(t, jsonOutput, "production")
}

func TestDataSourceProcessorReroute_MinimalConfig(t *testing.T) {
	resource := DataSourceProcessorReroute()

	// Test with just a destination
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"destination": "minimal-index",
	})

	ctx := context.Background()
	diags := resource.ReadContext(ctx, d, nil)

	assert.False(t, diags.HasError(), "Data source read should not have errors")
	assert.NotEmpty(t, d.Get("json"))

	jsonOutput := d.Get("json").(string)
	assert.Contains(t, jsonOutput, "minimal-index")
	assert.Contains(t, jsonOutput, "ignore_failure")
}

func TestDataSourceProcessorReroute_AllFields(t *testing.T) {
	resource := DataSourceProcessorReroute()

	// Test with all optional fields
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"destination":    "all-fields-index",
		"dataset":        "metrics",
		"namespace":      "development",
		"description":    "Full processor test",
		"if":             "ctx.field != null",
		"ignore_failure": true,
		"tag":            "reroute-tag",
		"on_failure":     []any{`{"set": {"field": "error", "value": "reroute_failed"}}`},
	})

	ctx := context.Background()
	diags := resource.ReadContext(ctx, d, nil)

	assert.False(t, diags.HasError(), "Data source read should not have errors")
	assert.NotEmpty(t, d.Get("json"))

	jsonOutput := d.Get("json").(string)
	assert.Contains(t, jsonOutput, "all-fields-index")
	assert.Contains(t, jsonOutput, "metrics")
	assert.Contains(t, jsonOutput, "development")
	assert.Contains(t, jsonOutput, "Full processor test")
	assert.Contains(t, jsonOutput, "ctx.field != null")
	assert.Contains(t, jsonOutput, "reroute-tag")
	assert.Contains(t, jsonOutput, "on_failure")
}
