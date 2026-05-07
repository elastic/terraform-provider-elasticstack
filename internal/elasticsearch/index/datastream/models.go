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

package datastream

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data is the Plugin Framework model for the elasticstack_elasticsearch_data_stream resource.
type Data struct {
	entitycore.ElasticsearchConnectionField
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	TimestampField types.String `tfsdk:"timestamp_field"`
	Indices        types.List   `tfsdk:"indices"`
	Generation     types.Int64  `tfsdk:"generation"`
	Metadata       types.String `tfsdk:"metadata"`
	Status         types.String `tfsdk:"status"`
	Template       types.String `tfsdk:"template"`
	ILMPolicy      types.String `tfsdk:"ilm_policy"`
	Hidden         types.Bool   `tfsdk:"hidden"`
	System         types.Bool   `tfsdk:"system"`
	Replicated     types.Bool   `tfsdk:"replicated"`
}

// indexModel represents a backing index entry in the indices list.
type indexModel struct {
	IndexName types.String `tfsdk:"index_name"`
	IndexUUID types.String `tfsdk:"index_uuid"`
}

func (d Data) GetID() types.String         { return d.ID }
func (d Data) GetResourceID() types.String { return d.Name }
