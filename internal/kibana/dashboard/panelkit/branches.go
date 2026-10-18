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

package panelkit

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ByFieldAttributes returns the attributes for the by_field branch of a by_field/by_esql
// discriminated union: shared plus the required data_view_id / field_name discriminating
// identifiers. shared is extended in place and returned, so callers must pass a fresh map (e.g.
// the result of their own package-local sharedAttributes() helper) rather than a cached one.
func ByFieldAttributes(shared map[string]schema.Attribute) map[string]schema.Attribute {
	shared["data_view_id"] = schema.StringAttribute{
		MarkdownDescription: "The ID of the data view that the control is tied to.",
		Required:            true,
	}
	shared["field_name"] = schema.StringAttribute{
		MarkdownDescription: "The name of the field in the data view that the control is tied to.",
		Required:            true,
	}
	return shared
}

// ByEsqlAttributes returns the attributes for the by_esql branch of a by_field/by_esql
// discriminated union: shared plus the required esql_query / values_source discriminating
// identifiers, with values_source restricted to EsqlValuesSourceUserValue. shared is extended in
// place and returned, so callers must pass a fresh map. esqlQueryDescription and
// valuesSourceDescription let each panel keep its own markdown wording for these two attributes
// while sharing the rest of the branch structure.
func ByEsqlAttributes(shared map[string]schema.Attribute, esqlQueryDescription, valuesSourceDescription string) map[string]schema.Attribute {
	shared["esql_query"] = schema.StringAttribute{
		MarkdownDescription: esqlQueryDescription,
		Required:            true,
	}
	shared["values_source"] = schema.StringAttribute{
		MarkdownDescription: valuesSourceDescription,
		Required:            true,
		Validators: []validator.String{
			stringvalidator.OneOf(EsqlValuesSourceUserValue),
		},
	}
	return shared
}
