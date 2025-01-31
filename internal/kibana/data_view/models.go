package data_view

import (
	"context"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *dataViewModel) populateFromAPI(ctx context.Context, data *kbapi.DataViewsDataViewResponseObject) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	resourceID := clients.CompositeId{
		ClusterId:  model.SpaceID.ValueString(),
		ResourceId: *data.DataView.Id,
	}

	// An existing null map should should be semantically equal to an empty map.
	semanticEqualEmptyMap := func(existing types.Map, incoming types.Map) types.Map {
		if !utils.IsKnown(existing) && len(incoming.Elements()) == 0 {
			return types.MapNull(incoming.ElementType(ctx))
		}
		return incoming
	}

	// An existing null slice should be semantically equal to an empty slice.
	semanticEqualEmptySlice := func(existing types.List, incoming types.List) types.List {
		if !utils.IsKnown(existing) && len(incoming.Elements()) == 0 {
			return types.ListNull(incoming.ElementType(ctx))
		}
		return incoming
	}

	handleNamespaces := func(existingList types.List, incoming []string) types.List {
		p := path.Root("data_view").AtName("namespaces")
		existing := utils.ListTypeToSlice_String(ctx, existingList, p, &diags)

		// incoming is typically [] except in <= v8.8.2 where it is omitted
		if incoming == nil {
			if existing == nil {
				return types.ListNull(types.StringType)
			} else {
				return existingList
			}
		}

		// An existing null slice should be semantically equal to a slice only containing the current space ID.
		if existing == nil && len(incoming) == 1 && incoming[0] == model.SpaceID.ValueString() {
			return types.ListNull(types.StringType)
		}

		// Keep the original ordering if equal but unordered
		// The API response is ordered by name.
		if len(existing) == len(incoming) {
			useExisting := true
			for _, ns := range existing {
				if !slices.Contains(incoming, ns) {
					useExisting = false
					break
				}
			}
			if useExisting {
				return existingList
			}
		}

		return utils.SliceToListType_String(ctx, incoming, p, &diags)
	}

	model.ID = types.StringValue(resourceID.String())
	model.DataView = utils.StructToObjectType(ctx, data.DataView, getDataViewAttrTypes(), path.Root("data_view"), &diags,
		func(item kbapi.DataViewsDataViewResponseObjectInner, meta utils.ObjectMeta) innerModel {
			dvInner := utils.ObjectTypeAs[innerModel](ctx, model.DataView, meta.Path, &diags)
			if dvInner == nil {
				dvInner = &innerModel{}
			}

			return innerModel{
				Title:         types.StringPointerValue(item.Title),
				Name:          types.StringPointerValue(item.Name),
				ID:            types.StringPointerValue(item.Id),
				TimeFieldName: types.StringPointerValue(item.TimeFieldName),
				SourceFilters: semanticEqualEmptySlice(dvInner.SourceFilters,
					utils.SliceToListType(ctx, utils.Deref(item.SourceFilters), types.StringType, meta.Path.AtName("source_filters"), &diags,
						func(item kbapi.DataViewsSourcefilterItem, meta utils.ListMeta) string {
							return item.Value
						})),
				FieldAttributes: semanticEqualEmptyMap(dvInner.FieldAttributes,
					utils.MapToMapType(ctx, utils.Deref(item.FieldAttrs), getFieldAttrElemType(), meta.Path.AtName("field_attrs"), &diags,
						func(item kbapi.DataViewsFieldattrs, meta utils.MapMeta) fieldAttrModel {
							return fieldAttrModel{
								CustomLabel: types.StringPointerValue(item.CustomLabel),
								Count:       types.Int64PointerValue(utils.Itol(item.Count)),
							}
						})),
				RuntimeFieldMap: semanticEqualEmptyMap(dvInner.RuntimeFieldMap,
					utils.MapToMapType(ctx, utils.Deref(item.RuntimeFieldMap), getRuntimeFieldMapElemType(), meta.Path.AtName("runtime_field_map"), &diags,
						func(item kbapi.DataViewsRuntimefieldmap, meta utils.MapMeta) runtimeFieldModel {
							return runtimeFieldModel{
								Type:         types.StringValue(item.Type),
								ScriptSource: types.StringPointerValue(item.Script.Source),
							}
						})),
				FieldFormats: semanticEqualEmptyMap(dvInner.FieldFormats,
					utils.MapToMapType(ctx, utils.Deref(item.FieldFormats), getFieldFormatElemType(), meta.Path.AtName("field_formats"), &diags,
						func(item kbapi.DataViewsFieldformat, meta utils.MapMeta) fieldFormatModel {
							return fieldFormatModel{
								ID: types.StringPointerValue(item.Id),
								Params: utils.StructToObjectType(ctx, item.Params, getFieldFormatParamsAttrTypes(), meta.Path.AtName("params"), &diags,
									func(item kbapi.DataViewsFieldformatParams, meta utils.ObjectMeta) fieldFormatParamsModel {
										return fieldFormatParamsModel{
											Pattern:                types.StringPointerValue(item.Pattern),
											UrlTemplate:            types.StringPointerValue(item.UrlTemplate),
											LabelTemplate:          types.StringPointerValue(item.LabelTemplate),
											InputFormat:            types.StringPointerValue(item.InputFormat),
											OutputFormat:           types.StringPointerValue(item.OutputFormat),
											OutputPrecision:        types.Int64PointerValue(utils.Itol(item.OutputPrecision)),
											IncludeSpaceWithSuffix: types.BoolPointerValue(item.IncludeSpaceWithSuffix),
											UseShortSuffix:         types.BoolPointerValue(item.UseShortSuffix),
											Timezone:               types.StringPointerValue(item.Timezone),
											FieldType:              types.StringPointerValue(item.FieldType),
											Colors: utils.SliceToListType(ctx, utils.Deref(item.Colors), getFieldFormatParamsColorsElemType(), meta.Path.AtName("colors"), meta.Diags,
												func(item kbapi.DataViewsFieldformatParamsColor, meta utils.ListMeta) colorConfigModel {
													return colorConfigModel{
														Range:      types.StringPointerValue(item.Range),
														Regex:      types.StringPointerValue(item.Regex),
														Text:       types.StringPointerValue(item.Text),
														Background: types.StringPointerValue(item.Background),
													}
												}),
											FieldLength: types.Int64PointerValue(utils.Itol(item.FieldLength)),
											Transform:   types.StringPointerValue(item.Transform),
											LookupEntries: utils.SliceToListType(ctx, utils.Deref(item.LookupEntries), getFieldFormatParamsLookupEntryElemType(), meta.Path.AtName("lookup_entries"), meta.Diags,
												func(item kbapi.DataViewsFieldformatParamsLookup, meta utils.ListMeta) lookupEntryModel {
													return lookupEntryModel{
														Key:   types.StringPointerValue(item.Key),
														Value: types.StringPointerValue(item.Value),
													}
												}),
											UnknownKeyValue: types.StringPointerValue(item.UnknownKeyValue),
											Type:            types.StringPointerValue(item.Type),
											Width:           types.Int64PointerValue(utils.Itol(item.Width)),
											Height:          types.Int64PointerValue(utils.Itol(item.Height)),
										}
									}),
							}
						})),
				AllowNoIndex: types.BoolPointerValue(item.AllowNoIndex),
				Namespaces:   handleNamespaces(dvInner.Namespaces, utils.Deref(item.Namespaces)),
			}
		})

	return diags
}

func (model dataViewModel) toAPICreateModel(ctx context.Context) (kbapi.DataViewsCreateDataViewRequestObject, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.DataViewsCreateDataViewRequestObject{
		DataView: *utils.ObjectTypeToStruct(ctx, model.DataView, path.Root("data_view"), &diags,
			func(item innerModel, meta utils.ObjectMeta) kbapi.DataViewsCreateDataViewRequestObjectInner {
				// Add spaceID to namespaces if missing
				spaceID := model.SpaceID.ValueString()
				namespaces := utils.ListTypeToSlice_String(ctx, item.Namespaces, meta.Path.AtName("namespaces"), &diags)
				if namespaces != nil && !slices.Contains(namespaces, spaceID) {
					namespaces = append(namespaces, spaceID)
				}

				return kbapi.DataViewsCreateDataViewRequestObjectInner{
					AllowNoIndex: item.AllowNoIndex.ValueBoolPointer(),
					FieldAttrs: utils.MapRef(utils.MapTypeToMap(ctx, item.FieldAttributes, meta.Path.AtName("field_attrs"), &diags,
						func(item fieldAttrModel, meta utils.MapMeta) kbapi.DataViewsFieldattrs {
							return kbapi.DataViewsFieldattrs{
								Count:       utils.Ltoi(item.Count.ValueInt64Pointer()),
								CustomLabel: item.CustomLabel.ValueStringPointer(),
							}
						})),
					FieldFormats:    utils.MapRef(convertFieldFormats(utils.MapTypeToMap(ctx, item.FieldFormats, meta.Path.AtName("field_formats"), &diags, convertFieldFormat))),
					Id:              utils.ValueStringPointer(item.ID),
					Name:            utils.ValueStringPointer(item.Name),
					Namespaces:      utils.SliceRef(namespaces),
					RuntimeFieldMap: utils.MapRef(utils.MapTypeToMap(ctx, item.RuntimeFieldMap, meta.Path.AtName("runtime_field_map"), &diags, convertRuntimeFieldMap)),
					SourceFilters:   utils.SliceRef(utils.ListTypeToSlice(ctx, item.SourceFilters, meta.Path.AtName("source_filters"), &diags, convertSourceFilter)),
					TimeFieldName:   utils.ValueStringPointer(item.TimeFieldName),
					Title:           item.Title.ValueString(),
				}
			}),
		Override: model.Override.ValueBoolPointer(),
	}

	return body, diags
}

func (model dataViewModel) toAPIUpdateModel(ctx context.Context) (kbapi.DataViewsUpdateDataViewRequestObject, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.DataViewsUpdateDataViewRequestObject{
		DataView: utils.Deref(utils.ObjectTypeToStruct(ctx, model.DataView, path.Root("data_view"), &diags,
			func(item innerModel, meta utils.ObjectMeta) kbapi.DataViewsUpdateDataViewRequestObjectInner {
				return kbapi.DataViewsUpdateDataViewRequestObjectInner{
					AllowNoIndex:    item.AllowNoIndex.ValueBoolPointer(),
					FieldFormats:    utils.MapRef(convertFieldFormats(utils.MapTypeToMap(ctx, item.FieldFormats, meta.Path.AtName("field_formats"), &diags, convertFieldFormat))),
					Name:            utils.ValueStringPointer(item.Name),
					RuntimeFieldMap: utils.MapRef(utils.MapTypeToMap(ctx, item.RuntimeFieldMap, meta.Path.AtName("runtime_field_map"), &diags, convertRuntimeFieldMap)),
					SourceFilters:   utils.SliceRef(utils.ListTypeToSlice(ctx, item.SourceFilters, meta.Path.AtName("source_filters"), &diags, convertSourceFilter)),
					TimeFieldName:   utils.ValueStringPointer(item.TimeFieldName),
					Title:           item.Title.ValueStringPointer(),
				}
			})),
	}

	return body, diags
}

func convertFieldFormats(src map[string]kbapi.DataViewsFieldformat) kbapi.DataViewsFieldformats {
	if src == nil {
		return nil
	}

	dst := make(kbapi.DataViewsFieldformats, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func convertFieldFormat(item fieldFormatModel, meta utils.MapMeta) kbapi.DataViewsFieldformat {
	return kbapi.DataViewsFieldformat{
		Id: item.ID.ValueStringPointer(),
		Params: utils.ObjectTypeToStruct(meta.Context, item.Params, meta.Path.AtName("params"), meta.Diags,
			func(item fieldFormatParamsModel, meta utils.ObjectMeta) kbapi.DataViewsFieldformatParams {
				return kbapi.DataViewsFieldformatParams{
					LabelTemplate:          item.LabelTemplate.ValueStringPointer(),
					Pattern:                item.Pattern.ValueStringPointer(),
					UrlTemplate:            item.UrlTemplate.ValueStringPointer(),
					InputFormat:            item.InputFormat.ValueStringPointer(),
					OutputFormat:           item.OutputFormat.ValueStringPointer(),
					OutputPrecision:        utils.Ltoi(item.OutputPrecision.ValueInt64Pointer()),
					IncludeSpaceWithSuffix: item.IncludeSpaceWithSuffix.ValueBoolPointer(),
					UseShortSuffix:         item.UseShortSuffix.ValueBoolPointer(),
					Timezone:               item.Timezone.ValueStringPointer(),
					FieldType:              item.FieldType.ValueStringPointer(),
					Colors: utils.SliceRef(utils.ListTypeToSlice(meta.Context, item.Colors, meta.Path.AtName("colors"), meta.Diags,
						func(item colorConfigModel, meta utils.ListMeta) kbapi.DataViewsFieldformatParamsColor {
							return kbapi.DataViewsFieldformatParamsColor{
								Background: item.Background.ValueStringPointer(),
								Range:      item.Range.ValueStringPointer(),
								Regex:      item.Regex.ValueStringPointer(),
								Text:       item.Text.ValueStringPointer(),
							}
						})),
					FieldLength: utils.Ltoi(item.FieldLength.ValueInt64Pointer()),
					Transform:   item.Transform.ValueStringPointer(),
					LookupEntries: utils.SliceRef(utils.ListTypeToSlice(meta.Context, item.LookupEntries, meta.Path.AtName("lookup_entries"), meta.Diags,
						func(item lookupEntryModel, meta utils.ListMeta) kbapi.DataViewsFieldformatParamsLookup {
							return kbapi.DataViewsFieldformatParamsLookup{
								Key:   item.Key.ValueStringPointer(),
								Value: item.Value.ValueStringPointer(),
							}
						})),
					UnknownKeyValue: item.UnknownKeyValue.ValueStringPointer(),
					Type:            item.Type.ValueStringPointer(),
					Width:           utils.Ltoi(item.Width.ValueInt64Pointer()),
					Height:          utils.Ltoi(item.Height.ValueInt64Pointer()),
				}
			}),
	}
}

func convertRuntimeFieldMap(item runtimeFieldModel, meta utils.MapMeta) kbapi.DataViewsRuntimefieldmap {
	return kbapi.DataViewsRuntimefieldmap{
		Type: item.Type.ValueString(),
		Script: kbapi.DataViewsRuntimefieldmapScript{
			Source: item.ScriptSource.ValueStringPointer(),
		},
	}
}

func convertSourceFilter(item string, meta utils.ListMeta) kbapi.DataViewsSourcefilterItem {
	return kbapi.DataViewsSourcefilterItem{Value: item}
}

func (model dataViewModel) getViewIDAndSpaceID() (viewID string, spaceID string) {
	viewID = model.ID.ValueString()
	spaceID = model.SpaceID.ValueString()

	resourceID := model.ID.ValueString()
	maybeCompositeID, _ := clients.CompositeIdFromStr(resourceID)
	if maybeCompositeID != nil {
		viewID = maybeCompositeID.ResourceId
		spaceID = maybeCompositeID.ClusterId
	}

	return
}

type dataViewModel struct {
	ID       types.String `tfsdk:"id"`
	SpaceID  types.String `tfsdk:"space_id"`
	Override types.Bool   `tfsdk:"override"`
	DataView types.Object `tfsdk:"data_view"` //> innerModel
}

type innerModel struct {
	Title           types.String `tfsdk:"title"`
	Name            types.String `tfsdk:"name"`
	ID              types.String `tfsdk:"id"`
	TimeFieldName   types.String `tfsdk:"time_field_name"`
	SourceFilters   types.List   `tfsdk:"source_filters"`    //> string
	FieldAttributes types.Map    `tfsdk:"field_attrs"`       //> fieldAttrModel
	RuntimeFieldMap types.Map    `tfsdk:"runtime_field_map"` //> runtimeFieldModel
	FieldFormats    types.Map    `tfsdk:"field_formats"`     //> fieldFormatModel
	AllowNoIndex    types.Bool   `tfsdk:"allow_no_index"`
	Namespaces      types.List   `tfsdk:"namespaces"` //> string
}

type fieldAttrModel struct {
	CustomLabel types.String `tfsdk:"custom_label"`
	Count       types.Int64  `tfsdk:"count"`
}

type runtimeFieldModel struct {
	Type         types.String `tfsdk:"type"`
	ScriptSource types.String `tfsdk:"script_source"`
}

type fieldFormatModel struct {
	ID     types.String `tfsdk:"id"`
	Params types.Object `tfsdk:"params"` //> fieldFormatParamsModel
}

type fieldFormatParamsModel struct {
	Pattern                types.String `tfsdk:"pattern"`
	UrlTemplate            types.String `tfsdk:"urltemplate"`
	LabelTemplate          types.String `tfsdk:"labeltemplate"`
	InputFormat            types.String `tfsdk:"input_format"`
	OutputFormat           types.String `tfsdk:"output_format"`
	OutputPrecision        types.Int64  `tfsdk:"output_precision"`
	IncludeSpaceWithSuffix types.Bool   `tfsdk:"include_space_with_suffix"`
	UseShortSuffix         types.Bool   `tfsdk:"use_short_suffix"`
	Timezone               types.String `tfsdk:"timezone"`
	FieldType              types.String `tfsdk:"field_type"`
	Colors                 types.List   `tfsdk:"colors"` //> colorConfigModel
	FieldLength            types.Int64  `tfsdk:"field_length"`
	Transform              types.String `tfsdk:"transform"`
	LookupEntries          types.List   `tfsdk:"lookup_entries"` //> lookupEntryModel
	UnknownKeyValue        types.String `tfsdk:"unknown_key_value"`
	Type                   types.String `tfsdk:"type"`
	Width                  types.Int64  `tfsdk:"width"`
	Height                 types.Int64  `tfsdk:"height"`
}

type colorConfigModel struct {
	Range      types.String `tfsdk:"range"`
	Regex      types.String `tfsdk:"regex"`
	Text       types.String `tfsdk:"text"`
	Background types.String `tfsdk:"background"`
}

type lookupEntryModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}
