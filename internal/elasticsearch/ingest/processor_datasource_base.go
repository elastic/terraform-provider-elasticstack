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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ProcessorModel is the interface that all ingest processor data source models
// must implement to be used with the generic processorDataSource.
type ProcessorModel interface {
	TypeName() string
	MarshalBody() (any, diag.Diagnostics)
	SetID(id string)
	SetJSON(json string)
}

// processorDataSource is a generic Plugin Framework data source implementation
// that eliminates per-processor Read duplication.
type processorDataSource[T ProcessorModel] struct {
	model  T
	schema schema.Schema
}

// compile-time interface checks with a concrete pointer type.
var (
	_ datasource.DataSource              = NewProcessorDataSource(&processorDropModel{}, schema.Schema{})
	_ datasource.DataSourceWithConfigure = NewProcessorDataSource(&processorDropModel{}, schema.Schema{}).(datasource.DataSourceWithConfigure)
)

// NewProcessorDataSource returns a new datasource.DataSource for the given
// processor model and schema. Each concrete processor calls this in its own
// constructor.
func NewProcessorDataSource[T ProcessorModel](model T, s schema.Schema) datasource.DataSource {
	return &processorDataSource[T]{model: model, schema: s}
}

// Metadata implements datasource.DataSource.
func (d *processorDataSource[T]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_ingest_processor_" + d.model.TypeName()
}

// Schema implements datasource.DataSource.
func (d *processorDataSource[T]) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schema
}

// Read implements datasource.DataSource.
func (d *processorDataSource[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	model := d.model

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := model.MarshalBody()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jsonStr, hash, diags := marshalAndHash(model.TypeName(), body)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	model.SetID(hash)
	model.SetJSON(jsonStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Configure implements datasource.DataSourceWithConfigure.
// No Elasticsearch connection is needed for processor data sources.
func (d *processorDataSource[T]) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	// no-op
}

// marshalAndHash wraps the processor body as {"typeName": body}, marshals it
// to indented JSON, computes a hash, and returns all three values.
func marshalAndHash(typeName string, body any) (jsonStr string, hash string, diags diag.Diagnostics) {
	wrapped := map[string]any{typeName: body}

	b, err := json.MarshalIndent(wrapped, "", " ")
	if err != nil {
		diags.AddError("Failed to marshal processor JSON", err.Error())
		return "", "", diags
	}

	jsonStr = string(b)
	hashPtr, err := typeutils.StringToHash(jsonStr)
	if err != nil {
		diags.AddError("Failed to hash processor JSON", err.Error())
		return "", "", diags
	}
	if hashPtr == nil {
		diags.AddError("Failed to hash processor JSON", "hash result is nil")
		return "", "", diags
	}

	return jsonStr, *hashPtr, diags
}

// IsKnown returns true if the value is not null and not unknown.
func IsKnown[T interface {
	IsNull() bool
	IsUnknown() bool
}](v T) bool {
	return !v.IsNull() && !v.IsUnknown()
}
