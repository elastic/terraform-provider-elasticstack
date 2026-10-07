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

package indexsettings

import (
	"context"

	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// getSchemaFactory returns the resource schema for
// elasticstack_elasticsearch_index_settings: the shared typed dynamic-setting
// attributes from indexparent.GetDynamicSettingAttributes() merged with the
// id, index and settings_json attributes. The provider envelope injects the
// elasticsearch_connection block and the timeouts attribute.
func getSchemaFactory(_ context.Context) schema.Schema {
	attributes := indexparent.GetDynamicSettingAttributes()

	attributes["id"] = schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Computed `<cluster_uuid>/<index_name>` identifier of the target index.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}

	attributes["index"] = schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Name of the target Elasticsearch index; changing it forces replacement. The name is treated as an already-resolved concrete index (no date-math resolution).",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}

	attributes["settings_json"] = schema.StringAttribute{
		Optional: true,
		MarkdownDescription: "Flat JSON object of dynamic index settings not covered by the typed attributes " +
			"(e.g. `{\"max_result_window\": 20000}`). Use flat dotted setting keys; creation-time-only " +
			"(static) settings, nested objects and explicit `null` values are rejected. To reset a setting, " +
			"omit it. Keys set here must not also be set via a typed attribute.",
		CustomType: jsontypes.NormalizedType{},
		Validators: []validator.String{
			validators.StringIsJSONObject{
				NonEmpty: true,
			},
			settingsJSONValidator{},
		},
	}

	return schema.Schema{
		MarkdownDescription: "Manages the dynamic settings of an existing Elasticsearch index without managing the index itself. " +
			"Changing or omitting a setting updates or resets it on the index; destroy does not revert settings " +
			"and the index is left in place.",
		Attributes: attributes,
	}
}
