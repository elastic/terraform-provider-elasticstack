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

package template

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// applyAllowCustomRouting8xWorkaround mirrors resourceIndexTemplatePut in the SDK: when prior state had
// allow_custom_routing=true and configuration does not explicitly set that attribute to true, re-send
// allow_custom_routing=false in the PUT body. toAPIModel only emits the field when true, so without this
// pass Elasticsearch 8.x can keep the old true value when practitioners remove the attribute from HCL.
func applyAllowCustomRouting8xWorkaround(ctx context.Context, prior, config Model, indexTemplate *models.IndexTemplate) {
	if !dataStreamAllowCustomRoutingWasTrue(ctx, prior.DataStream) {
		return
	}
	if configSetsAllowCustomRoutingTrue(config.DataStream) {
		return
	}
	f := false
	if indexTemplate.DataStream == nil {
		indexTemplate.DataStream = &models.DataStreamSettings{}
	}
	indexTemplate.DataStream.AllowCustomRouting = &f
}

func configSetsAllowCustomRoutingTrue(configDataStream types.Object) bool {
	if configDataStream.IsNull() || configDataStream.IsUnknown() {
		return false
	}
	v, ok := configDataStream.Attributes()[attrAllowCustomRouting]
	if !ok || v.IsNull() || v.IsUnknown() {
		return false
	}
	b, ok := v.(types.Bool)
	if !ok {
		return false
	}
	return b.ValueBool()
}

func dataStreamAllowCustomRoutingWasTrue(ctx context.Context, ds types.Object) bool {
	if ds.IsNull() || ds.IsUnknown() {
		return false
	}
	var m DataStreamModel
	diags := ds.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return false
	}
	return !m.AllowCustomRouting.IsNull() && !m.AllowCustomRouting.IsUnknown() && m.AllowCustomRouting.ValueBool()
}
