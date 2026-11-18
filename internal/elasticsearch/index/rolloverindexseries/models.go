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

package rolloverindexseries

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the Terraform state for elasticstack_elasticsearch_rollover_index_series.
type Model struct {
	entitycore.ElasticsearchConnectionField
	entitycore.ResourceTimeoutsField

	ID                 types.String `tfsdk:"id"`
	Prefix             types.String `tfsdk:"prefix"`
	RolloverAlias      types.String `tfsdk:"rollover_alias"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	WriteIndex         types.String `tfsdk:"write_index"`
	Indices            types.Set    `tfsdk:"indices"`
}

func (m Model) GetID() types.String { return m.ID }

func (m Model) GetResourceID() types.String { return m.RolloverAlias }

var _ entitycore.ElasticsearchResourceModel = Model{}

func (m Model) withObservedSeries(id string, series seriesView) Model {
	m.ID = types.StringValue(id)
	if series.HasWriter {
		m.WriteIndex = types.StringValue(series.Writer)
	} else {
		m.WriteIndex = types.StringNull()
	}
	elements := make([]attr.Value, 0, len(series.Members))
	for _, member := range series.Members {
		elements = append(elements, types.StringValue(member.Name))
	}
	m.Indices = types.SetValueMust(types.StringType, elements)
	return m
}
