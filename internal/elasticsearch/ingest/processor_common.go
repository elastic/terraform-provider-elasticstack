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
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CommonProcessorModel holds the PF schema state shared by every ingest
// processor data source. Embed this struct in concrete processor models.
type CommonProcessorModel struct {
	ID            types.String `tfsdk:"id"`
	JSON          types.String `tfsdk:"json"`
	Description   types.String `tfsdk:"description"`
	If            types.String `tfsdk:"if"`
	IgnoreFailure types.Bool   `tfsdk:"ignore_failure"`
	OnFailure     types.List   `tfsdk:"on_failure"`
	Tag           types.String `tfsdk:"tag"`
}

func (m *CommonProcessorModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *CommonProcessorModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

// WithTargetField holds the field attributes shared by processors that write to
// a target field.
type WithTargetField struct {
	Field       types.String `tfsdk:"field"`
	TargetField types.String `tfsdk:"target_field"`
}

func (m WithTargetField) toTargetFieldBody() WithTargetFieldBody {
	var body WithTargetFieldBody

	if typeutils.IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if typeutils.IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}

	return body
}

// WithIgnorableTargetField holds the target field attributes shared by
// processors that can ignore missing input fields.
type WithIgnorableTargetField struct {
	WithTargetField
	IgnoreMissing types.Bool `tfsdk:"ignore_missing"`
}

func (m *WithIgnorableTargetField) toIgnorableTargetFieldBody(defaultIgnoreMissing bool) WithIgnorableTargetFieldBody {
	body := WithIgnorableTargetFieldBody{
		WithTargetFieldBody: m.toTargetFieldBody(),
	}

	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		// Normalize computed defaults while building the body so state matches the JSON.
		m.IgnoreMissing = types.BoolValue(defaultIgnoreMissing)
		body.IgnoreMissing = defaultIgnoreMissing
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}

	return body
}

// simpleIgnorableTargetFieldModel is the shared Terraform state model for
// processors whose only non-common attributes are field, target_field, and
// ignore_missing (bytes, html_strip, lowercase, registered_domain, rename,
// trim, uppercase, urldecode).
type simpleIgnorableTargetFieldModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
}

// simpleProcessorConfig holds the per-processor strings needed to build a
// simpleIgnorableTargetFieldDataSource without repeating boilerplate.
type simpleProcessorConfig struct {
	typeName        string
	description     string
	fieldDesc       string
	targetFieldDesc string
	targetRequired  bool
}

// simpleIgnorableTargetFieldDataSource is a specialised datasource.DataSource
// for the 8 simple-transform processors. It keeps the type name outside the
// model so that Config.Get cannot zero it out.
type simpleIgnorableTargetFieldDataSource struct {
	cfg    simpleProcessorConfig
	schema schema.Schema
}

var _ datasource.DataSource = (*simpleIgnorableTargetFieldDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*simpleIgnorableTargetFieldDataSource)(nil)

func (d *simpleIgnorableTargetFieldDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_ingest_processor_" + d.cfg.typeName
}

func (d *simpleIgnorableTargetFieldDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.schema
}

func (d *simpleIgnorableTargetFieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model simpleIgnorableTargetFieldModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var diags diag.Diagnostics
	body := simpleIgnorableTargetFieldBody{}
	body.CommonProcessorBody, diags = model.toCommonProcessorBody()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body.WithIgnorableTargetFieldBody = model.toIgnorableTargetFieldBody(false)

	jsonStr, hash, diags := marshalAndHash(d.cfg.typeName, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.SetID(hash)
	model.SetJSON(jsonStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Configure implements datasource.DataSourceWithConfigure.
func (d *simpleIgnorableTargetFieldDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

// newSimpleIgnorableTargetFieldDataSource builds a PF data source for a
// processor whose schema is exactly: id, json, field, target_field,
// ignore_missing, plus CommonProcessorSchemaAttributes.
func newSimpleIgnorableTargetFieldDataSource(cfg simpleProcessorConfig) datasource.DataSource {
	targetAttr := schema.StringAttribute{
		Description: cfg.targetFieldDesc,
		Optional:    true,
	}
	if cfg.targetRequired {
		targetAttr.Required = true
		targetAttr.Optional = false
	}

	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: descIdentifier,
			Computed:    true,
		},
		attrJSON: schema.StringAttribute{
			Description: descJSONDataSource,
			Computed:    true,
		},
		attrField: schema.StringAttribute{
			Description: cfg.fieldDesc,
			Required:    true,
		},
		attrTargetField: targetAttr,
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return &simpleIgnorableTargetFieldDataSource{
		cfg: cfg,
		schema: schema.Schema{
			Description: cfg.description,
			Attributes:  attrs,
		},
	}
}

// CommonProcessorSchemaAttributes returns the schema attributes that are common
// to all ingest processor data sources.
func CommonProcessorSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Description: "Description of the processor.",
			Optional:    true,
		},
		"if": schema.StringAttribute{
			Description: "Conditionally execute the processor",
			Optional:    true,
		},
		"ignore_failure": schema.BoolAttribute{
			Description: "Ignore failures for the processor.",
			Optional:    true,
			Computed:    true,
		},
		attrOnFailure: schema.ListAttribute{
			Description: "Handle failures for the processor.",
			Optional:    true,
			ElementType: jsontypes.NormalizedType{},
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"tag": schema.StringAttribute{
			Description: "Identifier for the processor.",
			Optional:    true,
		},
	}
}
