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

package componenttemplate

import (
	"context"
	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data is the Terraform plan/state shape for the component template resource.
type Data struct {
	entitycore.ResourceTimeoutsField
	ID                      types.String         `tfsdk:"id"`
	Name                    types.String         `tfsdk:"name"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Template                types.Object         `tfsdk:"template"`
	Version                 types.Int64          `tfsdk:"version"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
}

func (d Data) GetID() types.String                    { return d.ID }
func (d Data) GetResourceID() types.String            { return d.Name }
func (d Data) GetElasticsearchConnection() types.List { return d.ElasticsearchConnection }

// GetTemplate satisfies [templateutil.TemplateModel].
func (d Data) GetTemplate() types.Object { return d.Template }

// WithTemplate satisfies [templateutil.TemplateModel].
func (d Data) WithTemplate(t types.Object) Data { d.Template = t; return d }

var _ entitycore.WithVersionRequirements = Data{}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (d Data) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	if d.Template.IsNull() || d.Template.IsUnknown() {
		return nil, nil
	}
	return datastreamoptions.GetVersionRequirements(d.Template)
}

// TemplateModel is the inner shape of the template block.
type TemplateModel struct {
	Alias             types.Set                      `tfsdk:"alias"`
	Mappings          esindex.MappingsValue          `tfsdk:"mappings"`
	Settings          customtypes.IndexSettingsValue `tfsdk:"settings"`
	DataStreamOptions types.Object                   `tfsdk:"data_stream_options"`
}
