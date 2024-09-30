package data_view

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/data_views"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages Kibana data views",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated ID for the data view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"override": schema.BoolAttribute{
				Description: "Overrides an existing data view if a data view with the provided title already exists.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"data_view": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						Description: "Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (*).",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"name": schema.StringAttribute{
						Description: "The Data view name.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"id": schema.StringAttribute{
						Description: "Saved object ID.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplace(),
						},
					},
					"time_field_name": schema.StringAttribute{
						Description: "Timestamp field name, which you use for time-based Data views.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"source_filters": schema.ListAttribute{
						Description: "List of field names you want to filter out in Discover.",
						ElementType: types.StringType,
						Optional:    true,
					},
					"field_attrs": schema.MapNestedAttribute{
						Description: "Map of field attributes by field name.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"custom_label": schema.StringAttribute{
									Description: "Custom label for the field.",
									Optional:    true,
								},
								"count": schema.Int64Attribute{
									Description: "Popularity count for the field.",
									Optional:    true,
								},
							},
						},
						Optional: true,
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.RequiresReplace(),
						},
					},
					"runtime_field_map": schema.MapNestedAttribute{
						Description: "Map of runtime field definitions by field name.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "Mapping type of the runtime field. For more information, check [Field data types](https://www.elastic.co/guide/en/elasticsearch/reference/8.11/mapping-types.html).",
									Required:            true,
								},
								"script_source": schema.StringAttribute{
									Description: "Script of the runtime field.",
									Required:    true,
								},
							},
						},
					},
					"field_formats": schema.MapNestedAttribute{
						Description: "Map of field formats by field name.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required: true,
								},
								"params": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"pattern": schema.StringAttribute{
											Optional: true,
										},
										"urltemplate": schema.StringAttribute{
											Optional: true,
										},
										"labeltemplate": schema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
					"allow_no_index": schema.BoolAttribute{
						Description: "Allows the Data view saved object to exist before the data is available.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"namespaces": schema.ListAttribute{
						Description: "Array of space IDs for sharing the Data view between multiple spaces.",
						ElementType: types.StringType,
						Optional:    true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

type Resource struct {
	client *clients.ApiClient
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_kibana_data_view"
}

type tfModelV0 struct {
	ID       types.String `tfsdk:"id"`
	SpaceID  types.String `tfsdk:"space_id"`
	Override types.Bool   `tfsdk:"override"`
	DataView types.Object `tfsdk:"data_view"` //> dataViewV0
}

type apiModelV0 struct {
	ID       string        `tfsdk:"id"`
	SpaceID  string        `tfsdk:"space_id"`
	Override bool          `tfsdk:"override"`
	DataView apiDataViewV0 `tfsdk:"data_view"`
}

func (m tfModelV0) ToCreateRequest(ctx context.Context) (data_views.CreateDataViewRequestObject, diag.Diagnostics) {
	apiModel := data_views.CreateDataViewRequestObject{
		Override: m.Override.ValueBoolPointer(),
	}

	var dataView tfDataViewV0
	if diags := m.DataView.As(ctx, &dataView, basetypes.ObjectAsOptions{}); diags.HasError() {
		return data_views.CreateDataViewRequestObject{}, diags
	}

	dv, diags := dataView.ToCreateRequest(ctx, m.SpaceID.ValueString())
	if diags.HasError() {
		return data_views.CreateDataViewRequestObject{}, diags
	}

	apiModel.DataView = dv
	return apiModel, nil
}

func (m tfModelV0) ToUpdateRequest(ctx context.Context) (data_views.UpdateDataViewRequestObject, diag.Diagnostics) {
	apiModel := data_views.UpdateDataViewRequestObject{}

	var dataView tfDataViewV0
	if diags := m.DataView.As(ctx, &dataView, basetypes.ObjectAsOptions{}); diags.HasError() {
		return data_views.UpdateDataViewRequestObject{}, diags
	}

	dv, diags := dataView.ToUpdateRequest(ctx)
	if diags.HasError() {
		return data_views.UpdateDataViewRequestObject{}, diags
	}

	apiModel.DataView = dv
	return apiModel, nil
}

func (m tfModelV0) FromResponse(ctx context.Context, resp *data_views.DataViewResponseObject) (apiModelV0, diag.Diagnostics) {
	dv := apiDataViewV0{}
	if resp.HasDataView() {
		dv = dataViewFromResponse(resp.GetDataView())
	}

	var dataView tfDataViewV0
	if !m.DataView.IsNull() && !m.DataView.IsUnknown() {
		if diags := m.DataView.As(ctx, &dataView, basetypes.ObjectAsOptions{}); diags.HasError() {
			return apiModelV0{}, diags
		}

		namespaces, diags := dataView.getNamespaces(ctx, nil)
		if diags.HasError() {
			return apiModelV0{}, diags
		}

		dv.Namespaces = namespaces
	}

	_, spaceID := m.getIDAndSpaceID()
	model := apiModelV0{
		ID:       m.ID.ValueString(),
		SpaceID:  spaceID,
		DataView: dv,
		Override: m.Override.ValueBool(),
	}
	return model, nil
}

func (model tfModelV0) getIDAndSpaceID() (string, string) {
	id := model.ID.ValueString()
	spaceID := model.SpaceID.ValueString()
	maybeCompositeID, _ := clients.CompositeIdFromStr(id)
	if maybeCompositeID != nil {
		id = maybeCompositeID.ResourceId
		spaceID = maybeCompositeID.ClusterId
	}

	return id, spaceID
}

type tfDataViewV0 struct {
	Title           types.String `tfsdk:"title"`
	Name            types.String `tfsdk:"name"`
	ID              types.String `tfsdk:"id"`
	TimeFieldName   types.String `tfsdk:"time_field_name"`
	SourceFilters   types.List   `tfsdk:"source_filters"`    //> string
	FieldAttributes types.Map    `tfsdk:"field_attrs"`       //> fieldAttrsV0
	RuntimeFieldMap types.Map    `tfsdk:"runtime_field_map"` //> runtimeFieldV0
	FieldFormats    types.Map    `tfsdk:"field_formats"`     //> fieldFormatV0
	AllowNoIndex    types.Bool   `tfsdk:"allow_no_index"`
	Namespaces      types.List   `tfsdk:"namespaces"`
}

type apiDataViewV0 struct {
	Title           *string                      `tfsdk:"title"`
	Name            *string                      `tfsdk:"name"`
	ID              string                       `tfsdk:"id"`
	TimeFieldName   *string                      `tfsdk:"time_field_name"`
	SourceFilters   []string                     `tfsdk:"source_filters"`
	FieldAttributes map[string]apiFieldAttrsV0   `tfsdk:"field_attrs"`
	RuntimeFieldMap map[string]apiRuntimeFieldV0 `tfsdk:"runtime_field_map"`
	FieldFormats    map[string]apiFieldFormat    `tfsdk:"field_formats"`
	AllowNoIndex    bool                         `tfsdk:"allow_no_index"`
	Namespaces      []string                     `tfsdk:"namespaces"`
}

func dataViewFromResponse(resp data_views.DataViewResponseObjectDataView) apiDataViewV0 {
	dv := apiDataViewV0{
		Title:         resp.Title,
		Name:          resp.Name,
		ID:            resp.GetId(),
		TimeFieldName: resp.TimeFieldName,
		AllowNoIndex:  resp.GetAllowNoIndex(),
	}

	if sourceFilters := resp.GetSourceFilters(); len(sourceFilters) > 0 {
		tfFilters := []string{}
		for _, filter := range sourceFilters {
			tfFilters = append(tfFilters, filter.GetValue())
		}

		dv.SourceFilters = tfFilters
	}

	fieldFormats := map[string]apiFieldFormat{}
	for field, format := range resp.GetFieldFormats() {
		formatMap := format.(map[string]interface{})
		apiFormat := apiFieldFormat{
			ID: formatMap["id"].(string),
		}

		if params, ok := formatMap["params"]; ok {
			if paramsMap, ok := params.(map[string]interface{}); ok {
				fieldFormatParams := apiFieldFormatParams{}
				if pattern, ok := paramsMap["pattern"]; ok {
					fieldFormatParams.Pattern = pattern.(string)
				}
				if urltemplate, ok := paramsMap["urlTemplate"]; ok {
					fieldFormatParams.Urltemplate = urltemplate.(string)
				}
				if labeltemplate, ok := paramsMap["labelTemplate"]; ok {
					fieldFormatParams.Labeltemplate = labeltemplate.(string)
				}
			}
		}

		fieldFormats[field] = apiFormat
	}

	if len(fieldFormats) > 0 {
		dv.FieldFormats = fieldFormats
	}

	fieldAttrs := map[string]apiFieldAttrsV0{}
	for field, attrs := range resp.GetFieldAttrs() {
		attrsMap := attrs.(map[string]interface{})
		apiAttrs := apiFieldAttrsV0{}
		if label, ok := attrsMap["customLabel"].(string); ok {
			apiAttrs.CustomLabel = &label
		}

		if count, ok := attrsMap["count"]; ok {
			var count64 int64
			switch c := count.(type) {
			case float64:
				count64 = int64(c)
			case int64:
				count64 = c
			}
			apiAttrs.Count = &count64
		}

		fieldAttrs[field] = apiAttrs
	}

	if len(fieldAttrs) > 0 {
		dv.FieldAttributes = fieldAttrs
	}

	runtimeFields := map[string]apiRuntimeFieldV0{}
	for field, runtimeDefn := range resp.GetRuntimeFieldMap() {
		runtimeMap := runtimeDefn.(map[string]interface{})
		apiField := apiRuntimeFieldV0{}
		if t, ok := runtimeMap["type"].(string); ok {
			apiField.Type = t
		}
		if script, ok := runtimeMap["script"].(map[string]interface{}); ok {
			apiField.ScriptSource = script["source"].(string)
		}

		runtimeFields[field] = apiField
	}

	if len(runtimeFields) > 0 {
		dv.RuntimeFieldMap = runtimeFields
	}

	return dv
}

func (m tfDataViewV0) ToCreateRequest(ctx context.Context, spaceID string) (data_views.CreateDataViewRequestObjectDataView, diag.Diagnostics) {
	apiModel := data_views.CreateDataViewRequestObjectDataView{
		Title: m.Title.ValueString(),
	}

	if utils.IsKnown(m.ID) {
		apiModel.Id = m.ID.ValueStringPointer()
	}

	if utils.IsKnown(m.Name) {
		apiModel.Name = m.Name.ValueStringPointer()
	}

	if utils.IsKnown(m.TimeFieldName) {
		apiModel.TimeFieldName = m.TimeFieldName.ValueStringPointer()
	}

	if utils.IsKnown(m.AllowNoIndex) {
		apiModel.AllowNoIndex = m.AllowNoIndex.ValueBoolPointer()
	}

	var sourceFilters []string
	if diags := m.SourceFilters.ElementsAs(ctx, &sourceFilters, true); diags.HasError() {
		return data_views.CreateDataViewRequestObjectDataView{}, diags
	}
	if sourceFilters != nil {
		apiFilters := []data_views.SourcefiltersInner{}
		for _, filter := range sourceFilters {
			apiFilters = append(apiFilters, data_views.SourcefiltersInner{
				Value: filter,
			})
		}
		apiModel.SourceFilters = apiFilters
	}

	fieldFormats, diags := tfFieldFormatsToAPI(ctx, m.FieldFormats)
	if diags.HasError() {
		return data_views.CreateDataViewRequestObjectDataView{}, diags
	}

	if fieldFormats != nil {
		apiModel.FieldFormats = fieldFormats
	}

	var tfFieldAttrs map[string]tfFieldAttrsV0
	if diags := m.FieldAttributes.ElementsAs(ctx, &tfFieldAttrs, true); diags.HasError() {
		return data_views.CreateDataViewRequestObjectDataView{}, diags
	}

	apiFieldAttrs := map[string]interface{}{}
	for field, attrs := range tfFieldAttrs {
		apiAttrs := fieldAttr{}
		if !attrs.CustomLabel.IsUnknown() {
			apiAttrs.CustomLabel = attrs.CustomLabel.ValueStringPointer()
		}

		if !attrs.Count.IsUnknown() {
			apiAttrs.Count = attrs.Count.ValueInt64Pointer()
		}

		apiFieldAttrs[field] = apiAttrs
	}

	if len(apiFieldAttrs) > 0 {
		apiModel.FieldAttrs = apiFieldAttrs
	}

	var runtimeFields map[string]tfRuntimeFieldV0
	if diags := m.RuntimeFieldMap.ElementsAs(ctx, &runtimeFields, true); diags.HasError() {
		return data_views.CreateDataViewRequestObjectDataView{}, diags
	}

	apiRuntimeFields := map[string]interface{}{}
	for field, defn := range runtimeFields {
		apiRuntimeFields[field] = runtimeField{
			Type: defn.Type.ValueString(),
			Script: runtimeFieldSource{
				Source: defn.ScriptSource.ValueString(),
			},
		}
	}
	if len(apiRuntimeFields) > 0 {
		apiModel.RuntimeFieldMap = apiRuntimeFields
	}

	namespaces, diags := m.getNamespaces(ctx, &spaceID)
	if diags.HasError() {
		return data_views.CreateDataViewRequestObjectDataView{}, diags
	}

	if len(namespaces) > 0 {
		apiModel.Namespaces = namespaces
	}

	return apiModel, nil
}

func (m tfDataViewV0) getNamespaces(ctx context.Context, spaceID *string) ([]string, diag.Diagnostics) {
	var namespaces []string
	if diags := m.Namespaces.ElementsAs(ctx, &namespaces, true); diags.HasError() {
		return nil, diags
	}

	if len(namespaces) == 0 || spaceID == nil {
		return namespaces, nil
	}

	includesSpaceID := false
	for _, namespace := range namespaces {
		if namespace == *spaceID {
			includesSpaceID = true
		}
	}

	if !includesSpaceID {
		namespaces = append(namespaces, *spaceID)
	}

	return namespaces, nil
}

func (m tfDataViewV0) ToUpdateRequest(ctx context.Context) (data_views.UpdateDataViewRequestObjectDataView, diag.Diagnostics) {
	apiModel := data_views.UpdateDataViewRequestObjectDataView{
		Title:         m.Title.ValueStringPointer(),
		Name:          m.Name.ValueStringPointer(),
		TimeFieldName: m.TimeFieldName.ValueStringPointer(),
		AllowNoIndex:  m.AllowNoIndex.ValueBoolPointer(),
	}

	var sourceFilters []string
	if diags := m.SourceFilters.ElementsAs(ctx, &sourceFilters, true); diags.HasError() {
		return data_views.UpdateDataViewRequestObjectDataView{}, diags
	}

	if len(sourceFilters) > 0 {
		apiFilters := []data_views.SourcefiltersInner{}
		for _, filter := range sourceFilters {
			apiFilters = append(apiFilters, data_views.SourcefiltersInner{
				Value: filter,
			})
		}
		apiModel.SourceFilters = apiFilters
	}

	fieldFormats, diags := tfFieldFormatsToAPI(ctx, m.FieldFormats)
	if diags.HasError() {
		return data_views.UpdateDataViewRequestObjectDataView{}, diags
	}
	if fieldFormats != nil {
		apiModel.FieldFormats = fieldFormats
	}

	var tfFieldAttrs map[string]tfFieldAttrsV0
	if diags := m.FieldAttributes.ElementsAs(ctx, &tfFieldAttrs, true); diags.HasError() {
		return data_views.UpdateDataViewRequestObjectDataView{}, diags
	}

	var runtimeFields map[string]tfRuntimeFieldV0
	if diags := m.RuntimeFieldMap.ElementsAs(ctx, &runtimeFields, true); diags.HasError() {
		return data_views.UpdateDataViewRequestObjectDataView{}, diags
	}

	apiRuntimeFields := map[string]interface{}{}
	for field, defn := range runtimeFields {
		apiRuntimeFields[field] = runtimeField{
			Type: defn.Type.ValueString(),
			Script: runtimeFieldSource{
				Source: defn.ScriptSource.ValueString(),
			},
		}
	}
	if len(apiRuntimeFields) > 0 {
		apiModel.RuntimeFieldMap = apiRuntimeFields
	}

	return apiModel, nil
}

func tfFieldFormatsToAPI(ctx context.Context, fieldFormats types.Map) (map[string]interface{}, diag.Diagnostics) {
	if fieldFormats.IsNull() || fieldFormats.IsUnknown() {
		return nil, nil
	}
	var tfFieldFormats map[string]types.Object
	if diags := fieldFormats.ElementsAs(ctx, &tfFieldFormats, true); diags.HasError() {
		return nil, diags
	}
	if len(tfFieldFormats) == 0 {
		return nil, nil
	}

	result := map[string]interface{}{}
	for field, format := range tfFieldFormats {
		var tfFormat tfFieldFormatV0
		if diags := tfsdk.ValueAs(ctx, format, &tfFormat); diags.HasError() {
			return nil, diags
		}

		var apiParams *apiFieldFormatParams
		if !tfFormat.Params.IsNull() && !tfFormat.Params.IsUnknown() {
			var tfParams tfFieldFormatParamsV0

			if diags := tfsdk.ValueAs(ctx, tfFormat.Params, &tfParams); diags.HasError() {
				return nil, diags
			}

			apiParams = &apiFieldFormatParams{
				Pattern:       tfParams.Pattern.ValueString(),
				Urltemplate:   tfParams.Urltemplate.ValueString(),
				Labeltemplate: tfParams.Labeltemplate.ValueString(),
			}
		}

		result[field] = apiFieldFormat{
			ID:     tfFormat.ID.ValueString(),
			Params: apiParams,
		}
	}

	return result, nil
}

type tfFieldAttrsV0 struct {
	CustomLabel types.String `tfsdk:"custom_label"`
	Count       types.Int64  `tfsdk:"count"`
}

type apiFieldAttrsV0 struct {
	CustomLabel *string `tfsdk:"custom_label"`
	Count       *int64  `tfsdk:"count"`
}

type tfRuntimeFieldV0 struct {
	Type         types.String `tfsdk:"type"`
	ScriptSource types.String `tfsdk:"script_source"`
}

type apiRuntimeFieldV0 struct {
	Type         string `tfsdk:"type"`
	ScriptSource string `tfsdk:"script_source"`
}

type fieldAttr struct {
	CustomLabel *string `json:"customLabel"`
	Count       *int64  `json:"count"`
}

type runtimeField struct {
	Type   string             `tfsdk:"type" json:"type"`
	Script runtimeFieldSource `tfsdk:"script" json:"script"`
}

type runtimeFieldSource struct {
	Source string `tfsdk:"source" json:"source"`
}

type tfFieldFormatV0 struct {
	ID     types.String `tfsdk:"id"`
	Params types.Object `tfsdk:"params"`
}

type apiFieldFormat struct {
	ID     string                `tfsdk:"id" json:"id"`
	Params *apiFieldFormatParams `tfsdk:"params" json:"params"`
}

type tfFieldFormatParamsV0 struct {
	Pattern       types.String `tfsdk:"pattern"`
	Urltemplate   types.String `tfsdk:"urltemplate"`
	Labeltemplate types.String `tfsdk:"labeltemplate"`
}

type apiFieldFormatParams struct {
	Pattern       string `tfsdk:"pattern" json:"pattern,omitempty"`
	Urltemplate   string `tfsdk:"urltemplate" json:"urlTemplate,omitempty"`
	Labeltemplate string `tfsdk:"labeltemplate" json:"labelTemplate,omitempty"`
}
