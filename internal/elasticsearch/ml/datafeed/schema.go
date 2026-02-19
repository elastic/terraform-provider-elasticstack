package datafeed

import (
	"context"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *datafeedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: schemaMarkdownDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"datafeed_id": schema.StringAttribute{
				MarkdownDescription: datafeedIDMarkdownDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$|^[a-z0-9]$`),
						"must contain lowercase alphanumeric characters (a-z and 0-9), hyphens, and underscores. "+
							"It must start and end with alphanumeric characters",
					),
				},
			},
			"job_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the anomaly detection job. The job must exist before creating the datafeed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"indices": schema.ListAttribute{
				MarkdownDescription: "An array of index names. Wildcards are supported. If any of the indices are in remote clusters, the machine learning nodes must have the `remote_cluster_client` role.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: queryMarkdownDescription,
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aggregations": schema.StringAttribute{
				MarkdownDescription: aggregationsMarkdownDescription,
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("script_fields").Expression()),
				},
			},
			"script_fields": schema.StringAttribute{
				MarkdownDescription: scriptFieldsMarkdownDescription,
				Optional:            true,
				CustomType:          customtypes.NewJSONWithDefaultsType(populateScriptFieldsDefaults),
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("aggregations").Expression()),
				},
			},
			"runtime_mappings": schema.StringAttribute{
				MarkdownDescription: "Specifies runtime fields for the datafeed search. This should be a JSON object representing the runtime field mappings.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"scroll_size": schema.Int64Attribute{
				MarkdownDescription: scrollSizeMarkdownDescription,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: frequencyMarkdownDescription,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 150s, 10m, 1h)"),
				},
			},
			"query_delay": schema.StringAttribute{
				MarkdownDescription: queryDelayMarkdownDescription,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 60s, 2m)"),
				},
			},
			"max_empty_searches": schema.Int64Attribute{
				MarkdownDescription: maxEmptySearchesMarkdownDescription,
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"chunking_config": schema.SingleNestedAttribute{
				MarkdownDescription: chunkingConfigMarkdownDescription,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						MarkdownDescription: chunkingModeMarkdownDescription,
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("auto", "manual", "off"),
						},
					},
					"time_span": schema.StringAttribute{
						MarkdownDescription: "The time span for each chunk. Only applicable and required when mode is `manual`. Must be a valid duration.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 1h, 1d)"),
							validators.AllowedIfDependentPathEquals(path.Root("chunking_config").AtName("mode"), "manual"),
						},
					},
				},
			},
			"delayed_data_check_config": schema.SingleNestedAttribute{
				MarkdownDescription: delayedDataCheckConfigMarkdownDescription,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Specifies whether the datafeed periodically checks for delayed data.",
						Required:            true,
					},
					"check_window": schema.StringAttribute{
						MarkdownDescription: checkWindowMarkdownDescription,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d+[nsumdh]$`), "must be a valid duration (e.g., 2h, 1d)"),
						},
					},
				},
			},
			"indices_options": schema.SingleNestedAttribute{
				MarkdownDescription: "Specifies index expansion options that are used during search.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"expand_wildcards": schema.ListAttribute{
						MarkdownDescription: expandWildcardsMarkdownDescription,
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.List{
							listvalidator.ValueStringsAre(
								stringvalidator.OneOf("all", "open", "closed", "hidden", "none"),
							),
						},
					},
					"ignore_unavailable": schema.BoolAttribute{
						MarkdownDescription: "If true, unavailable indices (missing or closed) are ignored.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"allow_no_indices": schema.BoolAttribute{
						MarkdownDescription: "If true, wildcard indices expressions that resolve into no concrete indices are ignored. This includes the `_all` string or when no indices are specified.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"ignore_throttled": schema.BoolAttribute{
						MarkdownDescription: "If true, concrete, expanded, or aliased indices are ignored when frozen. This setting is deprecated.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						DeprecationMessage: "This setting is deprecated and will be removed in a future version.",
					},
				},
			},
		},
	}
}

// GetChunkingConfigAttrTypes returns the attribute types for chunking_config
func GetChunkingConfigAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["chunking_config"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// GetDelayedDataCheckConfigAttrTypes returns the attribute types for delayed_data_check_config
func GetDelayedDataCheckConfigAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["delayed_data_check_config"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// GetIndicesOptionsAttrTypes returns the attribute types for indices_options
func GetIndicesOptionsAttrTypes() map[string]attr.Type {
	return GetSchema().Attributes["indices_options"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
