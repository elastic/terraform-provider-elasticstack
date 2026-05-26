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

package synonyms

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = newSynonymSetResource()
	_ resource.ResourceWithConfigure   = newSynonymSetResource()
	_ resource.ResourceWithImportState = newSynonymSetResource()
)

type synonymSetResource struct {
	*entitycore.ElasticsearchResource[SynonymSetData]
}

func newSynonymSetResource() *synonymSetResource {
	return &synonymSetResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource("synonym_set", entitycore.ElasticsearchResourceOptions[SynonymSetData]{
			Schema: schemaFactory,
			Create: upsertSynonymSet,
			Read:   readSynonymSet,
			Update: upsertSynonymSet,
			Delete: deleteSynonymSet,
		}),
	}
}

// NewSynonymSetResource returns a new synonym set resource for registration with the provider.
func NewSynonymSetResource() resource.Resource { return newSynonymSetResource() }

func (r *synonymSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// schemaFactory returns the schema for the synonym set resource. The
// elasticsearch_connection block is injected automatically by the envelope.
func schemaFactory(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: synonymSetResourceMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
			},
			"synonym_set_id": schema.StringAttribute{
				MarkdownDescription: "The name of the synonym set. Must be unique within the Elasticsearch cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"synonyms_set": schema.ListNestedAttribute{
				MarkdownDescription: "The list of synonym rules for this synonym set.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier for this synonym rule. When omitted, the provider generates a UUID.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						synonymsAttrName: schema.StringAttribute{
							MarkdownDescription: "The synonym rule in Solr format (e.g. `\"i-pod, i pod => ipod\"` or `\"universe, cosmos\"`).",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

// Ensure SynonymSetData satisfies the type constraint at compile time.
var _ entitycore.ElasticsearchResourceModel = SynonymSetData{}
