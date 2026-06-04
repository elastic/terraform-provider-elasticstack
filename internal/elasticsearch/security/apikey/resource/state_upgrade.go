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

package resource

import (
	"context"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/apikey"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func schemaWithConnection(ctx context.Context, version int64) schema.Schema {
	s := getSchema(version)
	blocks := make(map[string]schema.Block, len(s.Blocks)+1)
	maps.Copy(blocks, s.Blocks)
	blocks["elasticsearch_connection"] = providerschema.GetEsFWConnectionBlock()
	s.Blocks = blocks

	// The current TfModel embeds entitycore.ResourceTimeoutsField (the envelope
	// injects a "timeouts" attribute into the live schema). Prior-version state
	// written before timeouts existed has no such attribute, so the prior schema
	// must declare it for UpgradeState to decode prior state into TfModel without
	// a "Struct defines fields not found in object: timeouts" error. Missing
	// timeouts in the raw prior state decodes to null.
	attrs := make(map[string]schema.Attribute, len(s.Attributes)+1)
	maps.Copy(attrs, s.Attributes)
	attrs["timeouts"] = timeouts.AttributesAll(ctx)
	s.Attributes = attrs

	return s
}

func (r *Resource) UpgradeState(ctx context.Context) map[int64]fwresource.StateUpgrader {
	schema0 := schemaWithConnection(ctx, 0)
	schema1 := schemaWithConnection(ctx, 1)
	return map[int64]fwresource.StateUpgrader{
		0: {
			PriorSchema: &schema0,
			StateUpgrader: func(ctx context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
				var model apikey.TfModel
				resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if typeutils.IsKnown(model.Expiration) && model.Expiration.ValueString() == "" {
					model.Expiration = basetypes.NewStringNull()
				}

				resp.State.Set(ctx, model)
			},
		},
		1: {
			PriorSchema: &schema1,
			StateUpgrader: func(ctx context.Context, req fwresource.UpgradeStateRequest, resp *fwresource.UpgradeStateResponse) {
				var model apikey.TfModel
				resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
				if resp.Diagnostics.HasError() {
					return
				}

				model.Type = basetypes.NewStringValue(apikey.DefaultAPIKeyType)

				resp.State.Set(ctx, model)
			},
		},
	}
}
