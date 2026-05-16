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
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type processorRegisteredDomainModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
}

func (m *processorRegisteredDomainModel) TypeName() string { return "registered_domain" }

func (m *processorRegisteredDomainModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorRegisteredDomainBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	return body, diags
}

// NewProcessorRegisteredDomainDataSource returns a PF data source for the registered_domain processor.
func NewProcessorRegisteredDomainDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"field": schema.StringAttribute{
			Description: "Field containing the source FQDN.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "Object field containing extracted domain components. If an `<empty string>`, the processor adds components to the document's root.",
			Optional:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorRegisteredDomainModel{}, schema.Schema{
		Description: processorRegisteredDomainDataSourceDescription,
		Attributes:  attrs,
	})
}
