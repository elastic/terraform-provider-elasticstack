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

package transform

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	destinationIndexAllowedCharsError = "must contain lower case alphanumeric characters and selected punctuation, see the " +
		"[indices create API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html" +
		"#indices-create-api-path-params) for more details"

	// currentSchemaVersion is the resource schema version. Bump when the on-disk
	// state shape changes; add a corresponding entry in resource.UpgradeState.
	currentSchemaVersion int64 = 1
)

var (
	transformNameAllowedCharsRegexp   = regexp.MustCompile(`^[a-z0-9_-]+$`)
	transformNameStartEndRegexp       = regexp.MustCompile(`^[a-z0-9].*[a-z0-9]$`)
	destinationIndexLeadingCharRegexp = regexp.MustCompile(`^[^-_+]`)
	destinationIndexAllowedRegexp     = regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`)
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Version:             currentSchemaVersion,
		MarkdownDescription: transformDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the transform you wish to create.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						transformNameAllowedCharsRegexp,
						"must contain only lower case alphanumeric characters, hyphens, and underscores",
					),
					stringvalidator.RegexMatches(
						transformNameStartEndRegexp,
						"must start and end with a lowercase alphanumeric character",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Free text description of the transform.",
				Optional:            true,
			},
			"pivot": schema.StringAttribute{
				MarkdownDescription: "The pivot method transforms the data by aggregating and grouping it. JSON definition expected. Either 'pivot' or 'latest' must be present.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("pivot"),
						path.MatchRoot("latest"),
					}...),
				},
			},
			"latest": schema.StringAttribute{
				MarkdownDescription: "The latest method transforms the data by finding the latest document for each unique key. JSON definition expected. Either 'pivot' or 'latest' must be present.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("pivot"),
						path.MatchRoot("latest"),
					}...),
				},
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: "The interval between checks for changes in the source indices when the transform is running continuously. Defaults to `1m`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1m"),
				Validators: []validator.String{
					validators.ElasticDuration(),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Defines optional transform metadata.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"align_checkpoints": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the transform checkpoint ranges should be optimized for performance.",
				Optional:            true,
			},
			"dates_as_epoch_millis": schema.BoolAttribute{
				MarkdownDescription: "Defines if dates in the output should be written as ISO formatted string (default) or as millis since epoch.",
				Optional:            true,
			},
			"deduce_mappings": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the transform should deduce the destination index mappings from the transform config.",
				Optional:            true,
			},
			"docs_per_second": schema.Float64Attribute{
				MarkdownDescription: "Specifies a limit on the number of input documents per second. Default (unset) value disables throttling.",
				Optional:            true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0),
				},
			},
			"max_page_search_size": schema.Int64Attribute{
				MarkdownDescription: "Defines the initial page size to use for the composite aggregation for each checkpoint. Default is 500.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(10, 65536),
				},
			},
			"num_failure_retries": schema.Int64Attribute{
				MarkdownDescription: "Defines the number of retries on a recoverable failure before the transform task is marked as failed. " +
					"The default value is the cluster-level setting num_transform_failure_retries.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 100),
				},
			},
			"unattended": schema.BoolAttribute{
				MarkdownDescription: "In unattended mode, the transform retries indefinitely in case of an error which means the transform never fails.",
				Optional:            true,
			},
			"defer_validation": schema.BoolAttribute{
				MarkdownDescription: deferValidationDescription,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: timeoutDescription,
				Optional:            true,
				Computed:            true,
				CustomType:          customtypes.DurationType{},
				Default:             stringdefault.StaticString("30s"),
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Controls whether the transform should be started or stopped. Default is `false` (stopped).",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"source": schema.SingleNestedBlock{
				MarkdownDescription: "The source of the data for the transform.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"indices": schema.ListAttribute{
						MarkdownDescription: "The source indices for the transform.",
						Required:            true,
						ElementType:         stringAttributeType,
					},
					"query": schema.StringAttribute{
						MarkdownDescription: "A query clause that retrieves a subset of data from the source index.",
						Optional:            true,
						Computed:            true,
						CustomType:          jsontypes.NormalizedType{},
						Default:             stringdefault.StaticString(`{"match_all":{}}`),
					},
					"runtime_mappings": schema.StringAttribute{
						MarkdownDescription: "Definitions of search-time runtime fields that can be used by the transform.",
						Optional:            true,
						CustomType:          jsontypes.NormalizedType{},
					},
				},
			},
			"destination": schema.SingleNestedBlock{
				MarkdownDescription: "The destination for the transform.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"index": schema.StringAttribute{
						MarkdownDescription: "The destination index for the transform.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 255),
							stringvalidator.NoneOf(".", ".."),
							stringvalidator.RegexMatches(
								destinationIndexLeadingCharRegexp,
								"cannot start with -, _, +",
							),
							stringvalidator.RegexMatches(
								destinationIndexAllowedRegexp,
								destinationIndexAllowedCharsError,
							),
						},
					},
					"pipeline": schema.StringAttribute{
						MarkdownDescription: "The unique identifier for an ingest pipeline.",
						Optional:            true,
					},
				},
				Blocks: map[string]schema.Block{
					"aliases": schema.ListNestedBlock{
						MarkdownDescription: "The aliases that the destination index for the transform should have.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"alias": schema.StringAttribute{
									MarkdownDescription: "The name of the alias.",
									Required:            true,
								},
								"move_on_creation": schema.BoolAttribute{
									MarkdownDescription: "Whether the destination index should be the only index in this alias. Defaults to false.",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
							},
						},
					},
				},
			},
			"retention_policy": schema.SingleNestedBlock{
				MarkdownDescription: "Defines a retention policy for the transform.",
				Validators: []validator.Object{
					objectvalidator.AlsoRequires(path.MatchRelative().AtName("time")),
				},
				Blocks: map[string]schema.Block{
					"time": schema.SingleNestedBlock{
						MarkdownDescription: "Specifies that the transform uses a time field to set the retention policy.",
						Validators: []validator.Object{
							objectvalidator.AlsoRequires(
								path.MatchRelative().AtName("field"),
								path.MatchRelative().AtName("max_age"),
							),
						},
						Attributes: map[string]schema.Attribute{
							"field": schema.StringAttribute{
								MarkdownDescription: "The date field that is used to calculate the age of the document.",
								Optional:            true,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
							"max_age": schema.StringAttribute{
								MarkdownDescription: "Specifies the maximum age of a document in the destination index.",
								Optional:            true,
								Validators: []validator.String{
									validators.ElasticDuration(),
								},
							},
						},
					},
				},
			},
			"sync": schema.SingleNestedBlock{
				MarkdownDescription: "Defines the properties transforms require to run continuously.",
				Validators: []validator.Object{
					objectvalidator.AlsoRequires(path.MatchRelative().AtName("time")),
				},
				Blocks: map[string]schema.Block{
					"time": schema.SingleNestedBlock{
						MarkdownDescription: "Specifies that the transform uses a time field to synchronize the source and destination indices.",
						Validators: []validator.Object{
							objectvalidator.AlsoRequires(
								path.MatchRelative().AtName("field"),
							),
						},
						Attributes: map[string]schema.Attribute{
							"field": schema.StringAttribute{
								MarkdownDescription: "The date field that is used to identify new documents in the source.",
								Optional:            true,
								Validators: []validator.String{
									stringvalidator.LengthAtLeast(1),
								},
							},
							"delay": schema.StringAttribute{
								MarkdownDescription: "The time delay between the current time and the latest input data time. The default value is 60s.",
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("60s"),
								Validators: []validator.String{
									validators.ElasticDuration(),
								},
							},
						},
					},
				},
			},
		},
	}
}
