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

package ilm

import (
	"context"
	_ "embed"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed descriptions/ilm_resource.md
var resourceMarkdownDescription string

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             currentSchemaVersion,
		MarkdownDescription: resourceMarkdownDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(),
			ilmPhaseHot:                phaseHotBlock(),
			ilmPhaseWarm:               phaseWarmBlock(),
			ilmPhaseCold:               phaseColdBlock(),
			ilmPhaseFrozen:             phaseFrozenBlock(),
			ilmPhaseDelete:             phaseDeleteBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Identifier for the policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.StringAttribute{
				Description: "Optional user metadata about the ilm policy. Must be valid JSON document.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators:  []validator.String{esindex.StringIsJSONObject{}},
			},
			"modified_date": schema.StringAttribute{
				Description: "The DateTime of the last modification.",
				Computed:    true,
			},
		},
	}
}

func minAgeAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "ILM moves indices through the lifecycle according to their age. To control the timing of these transitions, you set a minimum age for each phase.",
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString("0ms"),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func phaseHotBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "The index is actively being updated and queried.",
		Attributes: map[string]schema.Attribute{
			"min_age": minAgeAttribute(),
		},
		Blocks: map[string]schema.Block{
			"set_priority":        blockSetPriority(),
			"unfollow":            blockUnfollow(),
			"rollover":            blockRollover(),
			"readonly":            blockReadonly(),
			"shrink":              blockShrink(),
			"forcemerge":          blockForcemerge(),
			"searchable_snapshot": blockSearchableSnapshot(),
			"downsample":          blockDownsample(),
		},
	}
}

func phaseWarmBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "The index is no longer being updated but is still being queried.",
		Attributes: map[string]schema.Attribute{
			"min_age": minAgeAttribute(),
		},
		Blocks: map[string]schema.Block{
			"set_priority": blockSetPriority(),
			"unfollow":     blockUnfollow(),
			"readonly":     blockReadonly(),
			"allocate":     blockAllocate(),
			"migrate":      blockMigrate(),
			"shrink":       blockShrink(),
			"forcemerge":   blockForcemerge(),
			"downsample":   blockDownsample(),
		},
	}
}

func phaseColdBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "The index is no longer being updated and is queried infrequently. The information still needs to be searchable, but it's okay if those queries are slower.",
		Attributes: map[string]schema.Attribute{
			"min_age": minAgeAttribute(),
		},
		Blocks: map[string]schema.Block{
			"set_priority":        blockSetPriority(),
			"unfollow":            blockUnfollow(),
			"readonly":            blockReadonly(),
			"searchable_snapshot": blockSearchableSnapshot(),
			"allocate":            blockAllocate(),
			"migrate":             blockMigrate(),
			"freeze":              blockFreeze(),
			"downsample":          blockDownsample(),
		},
	}
}

func phaseFrozenBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "The index is no longer being updated and is queried rarely. The information still needs to be searchable, but it's okay if those queries are extremely slow.",
		Attributes: map[string]schema.Attribute{
			"min_age": minAgeAttribute(),
		},
		Blocks: map[string]schema.Block{
			"searchable_snapshot": blockSearchableSnapshotInFrozenPhase(),
		},
		Validators: []validator.Object{
			objectvalidator.AlsoRequires(path.MatchRelative().AtName("searchable_snapshot")),
		},
	}
}

func phaseDeleteBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "The index is no longer needed and can safely be removed.",
		Attributes: map[string]schema.Attribute{
			"min_age": minAgeAttribute(),
		},
		Blocks: map[string]schema.Block{
			"wait_for_snapshot": blockWaitForSnapshot(),
			ilmPhaseDelete:      blockDeleteAction(),
		},
	}
}
