package data_view

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPopulateFromAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name          string
		response      kbapi.DataViewsDataViewResponseObject
		existingModel dataViewModel
		expectedModel dataViewModel
	}{
		{
			name: "all fields",
			existingModel: dataViewModel{
				ID:      types.StringValue("existing-space-id/existing-id"),
				SpaceID: types.StringValue("existing-space-id"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
					Namespaces:      utils.ListValueFrom(ctx, []string{"existing-namespace"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			response: kbapi.DataViewsDataViewResponseObject{
				DataView: &kbapi.DataViewsDataViewResponseObjectInner{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					Id:            utils.Pointer("id"),
					TimeFieldName: utils.Pointer("time_field_name"),
					AllowNoIndex:  utils.Pointer(true),
					SourceFilters: &kbapi.DataViewsSourcefilters{
						{Value: "field1"},
						{Value: "field2"},
					},
					FieldAttrs: &map[string]kbapi.DataViewsFieldattrs{
						"field1": {
							CustomLabel: utils.Pointer("custom_label"),
							Count:       utils.Pointer(10),
						},
					},
					FieldFormats: &kbapi.DataViewsFieldformats{
						"field1": kbapi.DataViewsFieldformat{
							Id:     utils.Pointer("field1"),
							Params: nil,
						},
					},
					RuntimeFieldMap: &map[string]kbapi.DataViewsRuntimefieldmap{
						"runtime_field": {
							Type: "keyword",
							Script: kbapi.DataViewsRuntimefieldmapScript{
								Source: utils.Pointer("emit('hello')"),
							},
						},
					},
				},
			},
			expectedModel: dataViewModel{
				ID:      types.StringValue("existing-space-id/id"),
				SpaceID: types.StringValue("existing-space-id"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					Title:         types.StringValue("title"),
					Name:          types.StringValue("name"),
					ID:            types.StringValue("id"),
					TimeFieldName: types.StringValue("time_field_name"),
					SourceFilters: utils.ListValueFrom(ctx, []string{"field1", "field2"}, types.StringType, path.Root("data_view").AtName("source_filters"), &diags),
					FieldAttributes: utils.MapValueFrom(ctx, map[string]fieldAttrModel{
						"field1": {
							CustomLabel: types.StringValue("custom_label"),
							Count:       types.Int64Value(10),
						},
					}, getFieldAttrElemType(), path.Root("data_view").AtName("field_attrs"), &diags),
					RuntimeFieldMap: utils.MapValueFrom(ctx, map[string]runtimeFieldModel{
						"runtime_field": {
							Type:         types.StringValue("keyword"),
							ScriptSource: types.StringValue("emit('hello')"),
						},
					}, getRuntimeFieldMapElemType(), path.Root("data_view").AtName("runtime_field_map"), &diags),
					FieldFormats: utils.MapValueFrom(ctx, map[string]fieldFormatModel{
						"field1": {
							ID:     types.StringValue("field1"),
							Params: types.ObjectNull(getFieldFormatParamsAttrTypes()),
						},
					}, getFieldFormatElemType(), path.Root("data_view").AtName("field_formats"), &diags),
					AllowNoIndex: types.BoolValue(true),
					Namespaces:   utils.ListValueFrom(ctx, []string{"existing-namespace"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
		},
		{
			// When sending no value, the response from Kibana is ["default"]
			name: "handleNamespaces_null_default",
			existingModel: dataViewModel{
				ID:      types.StringValue("default/id"),
				SpaceID: types.StringValue("default"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					ID:              types.StringValue("id"),
					Namespaces:      utils.ListValueFrom[string](ctx, nil, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			response: kbapi.DataViewsDataViewResponseObject{
				DataView: &kbapi.DataViewsDataViewResponseObjectInner{
					Id:         utils.Pointer("id"),
					Namespaces: &[]string{"default"},
				},
			},
			expectedModel: dataViewModel{
				ID:      types.StringValue("default/id"),
				SpaceID: types.StringValue("default"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					ID:              types.StringValue("id"),
					Namespaces:      utils.ListValueFrom[string](ctx, nil, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
		},
		{
			// When sending the SpaceID as the namespace, the response from Kibana should be the same
			name: "handleNamespaces_populated_default",
			existingModel: dataViewModel{
				ID:      types.StringValue("space_id/dataview_id"),
				SpaceID: types.StringValue("space_id"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					ID:              types.StringValue("dataview_id"),
					Namespaces:      utils.ListValueFrom(ctx, []string{"space_id"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			response: kbapi.DataViewsDataViewResponseObject{
				DataView: &kbapi.DataViewsDataViewResponseObjectInner{
					Id:         utils.Pointer("dataview_id"),
					Namespaces: &[]string{"space_id"},
				},
			},
			expectedModel: dataViewModel{
				ID:      types.StringValue("space_id/dataview_id"),
				SpaceID: types.StringValue("space_id"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					ID:              types.StringValue("dataview_id"),
					Namespaces:      utils.ListValueFrom(ctx, []string{"space_id"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
		},
		{
			// When sending a populated list, the response from Kibana should be the same list
			name: "handleNamespaces_populated_default",
			existingModel: dataViewModel{
				ID:      types.StringValue("test/placeholder"),
				SpaceID: types.StringValue("test"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					ID:              types.StringValue("placeholder"),
					Namespaces:      utils.ListValueFrom(ctx, []string{"ns1", "ns2"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			response: kbapi.DataViewsDataViewResponseObject{
				DataView: &kbapi.DataViewsDataViewResponseObjectInner{
					Id:         utils.Pointer("placeholder"),
					Namespaces: &[]string{"test", "ns1", "ns2"},
				},
			},
			expectedModel: dataViewModel{
				ID:      types.StringValue("test/placeholder"),
				SpaceID: types.StringValue("test"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					ID:              types.StringValue("placeholder"),
					Namespaces:      utils.ListValueFrom(ctx, []string{"ns1", "ns2"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := tt.existingModel.populateFromAPI(ctx, &tt.response, tt.existingModel.SpaceID.ValueString())

			require.Equal(t, tt.expectedModel, tt.existingModel)
			require.Empty(t, diags)
		})
	}
}

func TestToAPICreateModel(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name            string
		model           dataViewModel
		expectedRequest kbapi.DataViewsCreateDataViewRequestObject
	}{
		{
			name: "all fields",
			model: dataViewModel{
				SpaceID: types.StringValue("default"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					Title:         types.StringValue("title"),
					Name:          types.StringValue("name"),
					ID:            types.StringValue("id"),
					TimeFieldName: types.StringValue("time_field_name"),
					SourceFilters: utils.ListValueFrom(ctx, []string{"field1", "field2"}, types.StringType, path.Root("data_view").AtName("source_filters"), &diags),
					FieldAttributes: utils.MapValueFrom(ctx, map[string]fieldAttrModel{
						"field1": {
							CustomLabel: types.StringValue("custom_label"),
							Count:       types.Int64Value(10),
						},
					}, getFieldAttrElemType(), path.Root("data_view").AtName("field_attrs"), &diags),
					RuntimeFieldMap: utils.MapValueFrom(ctx, map[string]runtimeFieldModel{
						"runtime_field": {
							Type:         types.StringValue("keyword"),
							ScriptSource: types.StringValue("emit(\"hello\")"),
						},
					}, getRuntimeFieldMapElemType(), path.Root("data_view").AtName("runtime_field_map"), &diags),
					FieldFormats: utils.MapValueFrom(ctx, map[string]fieldFormatModel{
						"field1": {
							ID: types.StringValue("field1"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Pattern:       types.StringValue("0.00"),
								UrlTemplate:   types.StringValue("https://test.com/{{value}}"),
								LabelTemplate: types.StringValue("{{value}}"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("field1").AtName("params"), &diags),
						},
						"host.uptime": {
							ID: types.StringValue("duration"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								InputFormat:            types.StringValue("hours"),
								OutputFormat:           types.StringValue("humanizePrecise"),
								OutputPrecision:        types.Int64Value(2),
								IncludeSpaceWithSuffix: types.BoolValue(true),
								UseShortSuffix:         types.BoolValue(true),
								Colors:                 types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries:          types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("host.uptime").AtName("params"), &diags),
						},
						"user.last_password_change": {
							ID:     types.StringValue("relative_date"),
							Params: types.ObjectNull(getFieldFormatParamsAttrTypes()),
						},
						"user.last_login": {
							ID: types.StringValue("date"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Pattern:       types.StringValue("MMM D, YYYY @ HH:mm:ss.SSS"),
								Timezone:      types.StringValue("America/New_York"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.last_login").AtName("params"), &diags),
						},
						"user.is_active": {
							ID:     types.StringValue("boolean"),
							Params: types.ObjectNull(getFieldFormatParamsAttrTypes()),
						},
						"user.status": {
							ID: types.StringValue("color"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								FieldType: types.StringValue("string"),
								Colors: utils.ListValueFrom(ctx, []colorConfigModel{
									{
										Range:      types.StringValue("-Infinity:Infinity"),
										Regex:      types.StringValue("inactive*"),
										Text:       types.StringValue("#000000"),
										Background: types.StringValue("#ffffff"),
									},
								}, getFieldFormatParamsColorsElemType(), path.Root("data_view").AtName("field_formats").AtMapKey("user.status").AtName("params").AtName("colors"), &diags),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.status").AtName("params"), &diags),
						},
						"user.message": {
							ID: types.StringValue("truncate"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								FieldLength:   types.Int64Value(10),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.message").AtName("params"), &diags),
						},
						"host.name": {
							ID: types.StringValue("string"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Transform:     types.StringValue("upper"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("host.name").AtName("params"), &diags),
						},
						"response.code": {
							ID: types.StringValue("static_lookup"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Colors: types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: utils.ListValueFrom(ctx, []lookupEntryModel{
									{
										Key:   types.StringValue("200"),
										Value: types.StringValue("OK"),
									},
									{
										Key:   types.StringValue("404"),
										Value: types.StringValue("Not Found"),
									},
								}, getFieldFormatParamsLookupEntryElemType(), path.Root("data_view").AtName("field_formats").AtMapKey("response.code").AtName("params").AtName("lookup_entries"), &diags),
								UnknownKeyValue: types.StringValue("Unknown"),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("response.code").AtName("params"), &diags),
						},
						"url.original": {
							ID: types.StringValue("url"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Type:          types.StringValue("a"),
								UrlTemplate:   types.StringValue("URL TEMPLATE"),
								LabelTemplate: types.StringValue("LABEL TEMPLATE"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("url.original").AtName("params"), &diags),
						},
						"user.profile_picture": {
							ID: types.StringValue("url"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Type:          types.StringValue("img"),
								UrlTemplate:   types.StringValue("URL TEMPLATE"),
								LabelTemplate: types.StringValue("LABEL TEMPLATE"),
								Width:         types.Int64Value(6),
								Height:        types.Int64Value(4),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.profile_picture").AtName("params"), &diags),
						},
						"user.answering_message": {
							ID: types.StringValue("url"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Type:          types.StringValue("audio"),
								UrlTemplate:   types.StringValue("URL TEMPLATE"),
								LabelTemplate: types.StringValue("LABEL TEMPLATE"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.answering_message").AtName("params"), &diags),
						},
					}, getFieldFormatElemType(), path.Root("data_view").AtName("field_formats"), &diags),
					AllowNoIndex: types.BoolValue(true),
					Namespaces:   utils.ListValueFrom(ctx, []string{"backend", "o11y"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
				Override: types.BoolValue(true),
			},
			expectedRequest: kbapi.DataViewsCreateDataViewRequestObject{
				DataView: kbapi.DataViewsCreateDataViewRequestObjectInner{
					AllowNoIndex: utils.Pointer(true),
					FieldAttrs: &map[string]kbapi.DataViewsFieldattrs{
						"field1": {
							CustomLabel: utils.Pointer("custom_label"),
							Count:       utils.Pointer(10),
						},
					},
					FieldFormats: &kbapi.DataViewsFieldformats{
						"field1": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("field1"),
							Params: &kbapi.DataViewsFieldformatParams{
								Pattern:       utils.Pointer("0.00"),
								UrlTemplate:   utils.Pointer("https://test.com/{{value}}"),
								LabelTemplate: utils.Pointer("{{value}}"),
							},
						},
						"host.uptime": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("duration"),
							Params: &kbapi.DataViewsFieldformatParams{
								InputFormat:            utils.Pointer("hours"),
								OutputFormat:           utils.Pointer("humanizePrecise"),
								OutputPrecision:        utils.Pointer(2),
								IncludeSpaceWithSuffix: utils.Pointer(true),
								UseShortSuffix:         utils.Pointer(true),
							},
						},
						"user.last_password_change": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("relative_date"),
						},
						"user.last_login": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("date"),
							Params: &kbapi.DataViewsFieldformatParams{
								Pattern:  utils.Pointer("MMM D, YYYY @ HH:mm:ss.SSS"),
								Timezone: utils.Pointer("America/New_York"),
							},
						},
						"user.is_active": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("boolean"),
						},
						"user.status": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("color"),
							Params: &kbapi.DataViewsFieldformatParams{
								FieldType: utils.Pointer("string"),
								Colors: &[]kbapi.DataViewsFieldformatParamsColor{
									{
										Range:      utils.Pointer("-Infinity:Infinity"),
										Regex:      utils.Pointer("inactive*"),
										Text:       utils.Pointer("#000000"),
										Background: utils.Pointer("#ffffff"),
									},
								},
							},
						},
						"user.message": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("truncate"),
							Params: &kbapi.DataViewsFieldformatParams{
								FieldLength: utils.Pointer(10),
							},
						},
						"host.name": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("string"),
							Params: &kbapi.DataViewsFieldformatParams{
								Transform: utils.Pointer("upper"),
							},
						},
						"response.code": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("static_lookup"),
							Params: &kbapi.DataViewsFieldformatParams{
								LookupEntries: &[]kbapi.DataViewsFieldformatParamsLookup{
									{
										Key:   utils.Pointer("200"),
										Value: utils.Pointer("OK"),
									},
									{
										Key:   utils.Pointer("404"),
										Value: utils.Pointer("Not Found"),
									},
								},
								UnknownKeyValue: utils.Pointer("Unknown"),
							},
						},
						"url.original": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("url"),
							Params: &kbapi.DataViewsFieldformatParams{
								Type:          utils.Pointer("a"),
								UrlTemplate:   utils.Pointer("URL TEMPLATE"),
								LabelTemplate: utils.Pointer("LABEL TEMPLATE"),
							},
						},
						"user.profile_picture": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("url"),
							Params: &kbapi.DataViewsFieldformatParams{
								Type:          utils.Pointer("img"),
								UrlTemplate:   utils.Pointer("URL TEMPLATE"),
								LabelTemplate: utils.Pointer("LABEL TEMPLATE"),
								Width:         utils.Pointer(6),
								Height:        utils.Pointer(4),
							},
						},
						"user.answering_message": kbapi.DataViewsFieldformat{
							Id: utils.Pointer("url"),
							Params: &kbapi.DataViewsFieldformatParams{
								Type:          utils.Pointer("audio"),
								UrlTemplate:   utils.Pointer("URL TEMPLATE"),
								LabelTemplate: utils.Pointer("LABEL TEMPLATE"),
							},
						},
					},
					Id:         utils.Pointer("id"),
					Name:       utils.Pointer("name"),
					Namespaces: &[]string{"backend", "o11y", "default"},
					RuntimeFieldMap: &map[string]kbapi.DataViewsRuntimefieldmap{
						"runtime_field": {
							Type: "keyword",
							Script: kbapi.DataViewsRuntimefieldmapScript{
								Source: utils.Pointer("emit(\"hello\")"),
							},
						},
					},
					SourceFilters: &[]kbapi.DataViewsSourcefilterItem{
						{Value: "field1"},
						{Value: "field2"},
					},
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         "title",
				},
				Override: utils.Pointer(true),
			},
		},
		{
			name: "nil collections",
			model: dataViewModel{
				SpaceID: types.StringValue("default"),
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					Title:           types.StringValue("title"),
					Name:            types.StringValue("name"),
					ID:              types.StringValue("id"),
					TimeFieldName:   types.StringValue("time_field_name"),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
					AllowNoIndex:    types.BoolValue(true),
					Namespaces:      types.ListNull(types.StringType),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			expectedRequest: kbapi.DataViewsCreateDataViewRequestObject{
				Override: nil,
				DataView: kbapi.DataViewsCreateDataViewRequestObjectInner{
					AllowNoIndex:  utils.Pointer(true),
					Id:            utils.Pointer("id"),
					Name:          utils.Pointer("name"),
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         "title",
				},
			},
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, diags := tt.model.toAPICreateModel(ctx)
			require.Equal(t, tt.expectedRequest, request)
			require.Empty(t, diags)
		})
	}
}

func TestToAPIUpdateModel(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name            string
		model           dataViewModel
		expectedRequest kbapi.DataViewsUpdateDataViewRequestObject
	}{
		{
			name: "all fields",
			model: dataViewModel{
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					Title:         types.StringValue("title"),
					Name:          types.StringValue("name"),
					ID:            types.StringValue("id"),
					TimeFieldName: types.StringValue("time_field_name"),
					SourceFilters: utils.ListValueFrom(ctx, []string{"field1", "field2"}, types.StringType, path.Root("data_view").AtName("source_filters"), &diags),
					FieldAttributes: utils.MapValueFrom(ctx, map[string]fieldAttrModel{
						"field1": {
							CustomLabel: types.StringValue("custom_label"),
							Count:       types.Int64Value(10),
						},
					}, getFieldAttrElemType(), path.Root("data_view").AtName("field_attrs"), &diags),
					RuntimeFieldMap: utils.MapValueFrom(ctx, map[string]runtimeFieldModel{
						"runtime_field": {
							Type:         types.StringValue("keyword"),
							ScriptSource: types.StringValue("emit(\"hello\")"),
						},
					}, getRuntimeFieldMapElemType(), path.Root("data_view").AtName("runtime_field_map"), &diags),
					FieldFormats: utils.MapValueFrom(ctx, map[string]fieldFormatModel{
						"field1": {
							ID: types.StringValue("field1"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Pattern:       types.StringValue("0.00"),
								UrlTemplate:   types.StringValue("https://test.com/{{value}}"),
								LabelTemplate: types.StringValue("{{value}}"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("field1").AtName("params"), &diags),
						},
						"host.uptime": {
							ID: types.StringValue("duration"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								InputFormat:            types.StringValue("hours"),
								OutputFormat:           types.StringValue("humanizePrecise"),
								OutputPrecision:        types.Int64Value(2),
								IncludeSpaceWithSuffix: types.BoolValue(true),
								UseShortSuffix:         types.BoolValue(true),
								Colors:                 types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries:          types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("host.uptime").AtName("params"), &diags),
						},
						"user.last_password_change": {
							ID:     types.StringValue("relative_date"),
							Params: types.ObjectNull(getFieldFormatParamsAttrTypes()),
						},
						"user.last_login": {
							ID: types.StringValue("date"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Pattern:       types.StringValue("MMM D, YYYY @ HH:mm:ss.SSS"),
								Timezone:      types.StringValue("America/New_York"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.last_login").AtName("params"), &diags),
						},
						"user.is_active": {
							ID:     types.StringValue("boolean"),
							Params: types.ObjectNull(getFieldFormatParamsAttrTypes()),
						},
						"user.status": {
							ID: types.StringValue("color"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								FieldType: types.StringValue("string"),
								Colors: utils.ListValueFrom(ctx, []colorConfigModel{
									{
										Range:      types.StringValue("-Infinity:Infinity"),
										Regex:      types.StringValue("inactive*"),
										Text:       types.StringValue("#000000"),
										Background: types.StringValue("#ffffff"),
									},
								}, getFieldFormatParamsColorsElemType(), path.Root("data_view").AtName("field_formats").AtMapKey("user.status").AtName("params").AtName("colors"), &diags),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.status").AtName("params"), &diags),
						},
						"user.message": {
							ID: types.StringValue("truncate"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								FieldLength:   types.Int64Value(10),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.message").AtName("params"), &diags),
						},
						"host.name": {
							ID: types.StringValue("string"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Transform:     types.StringValue("upper"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("host.name").AtName("params"), &diags),
						},
						"response.code": {
							ID: types.StringValue("static_lookup"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								LookupEntries: utils.ListValueFrom(ctx, []lookupEntryModel{
									{
										Key:   types.StringValue("200"),
										Value: types.StringValue("OK"),
									},
									{
										Key:   types.StringValue("404"),
										Value: types.StringValue("Not Found"),
									},
								}, getFieldFormatParamsLookupEntryElemType(), path.Root("data_view").AtName("field_formats").AtMapKey("response.code").AtName("params").AtName("lookup_entries"), &diags),
								UnknownKeyValue: types.StringValue("Unknown"),
								Colors:          types.ListNull(getFieldFormatParamsColorsElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("response.code").AtName("params"), &diags),
						},
						"url.original": {
							ID: types.StringValue("url"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Type:          types.StringValue("a"),
								UrlTemplate:   types.StringValue("URL TEMPLATE"),
								LabelTemplate: types.StringValue("LABEL TEMPLATE"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("url.original").AtName("params"), &diags),
						},
						"user.profile_picture": {
							ID: types.StringValue("url"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Type:          types.StringValue("img"),
								UrlTemplate:   types.StringValue("URL TEMPLATE"),
								LabelTemplate: types.StringValue("LABEL TEMPLATE"),
								Width:         types.Int64Value(6),
								Height:        types.Int64Value(4),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.profile_picture").AtName("params"), &diags),
						},
						"user.answering_message": {
							ID: types.StringValue("url"),
							Params: utils.ObjectValueFrom(ctx, &fieldFormatParamsModel{
								Type:          types.StringValue("audio"),
								UrlTemplate:   types.StringValue("URL TEMPLATE"),
								LabelTemplate: types.StringValue("LABEL TEMPLATE"),
								Colors:        types.ListNull(getFieldFormatParamsColorsElemType()),
								LookupEntries: types.ListNull(getFieldFormatParamsLookupEntryElemType()),
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("user.answering_message").AtName("params"), &diags),
						},
					}, getFieldFormatElemType(), path.Root("data_view").AtName("field_formats"), &diags),
					AllowNoIndex: types.BoolValue(true),
					Namespaces:   utils.ListValueFrom(ctx, []string{"default", "o11y"}, types.StringType, path.Root("data_view").AtName("namespaces"), &diags),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			expectedRequest: kbapi.DataViewsUpdateDataViewRequestObject{
				DataView: kbapi.DataViewsUpdateDataViewRequestObjectInner{
					AllowNoIndex: utils.Pointer(true),
					FieldFormats: &kbapi.DataViewsFieldformats{
						"field1": {
							Id: utils.Pointer("field1"),
							Params: &kbapi.DataViewsFieldformatParams{
								Pattern:       utils.Pointer("0.00"),
								UrlTemplate:   utils.Pointer("https://test.com/{{value}}"),
								LabelTemplate: utils.Pointer("{{value}}"),
							},
						},
						"host.uptime": {
							Id: utils.Pointer("duration"),
							Params: &kbapi.DataViewsFieldformatParams{
								InputFormat:            utils.Pointer("hours"),
								OutputFormat:           utils.Pointer("humanizePrecise"),
								OutputPrecision:        utils.Pointer(2),
								IncludeSpaceWithSuffix: utils.Pointer(true),
								UseShortSuffix:         utils.Pointer(true),
							},
						},
						"user.last_password_change": {
							Id: utils.Pointer("relative_date"),
						},
						"user.last_login": {
							Id: utils.Pointer("date"),
							Params: &kbapi.DataViewsFieldformatParams{
								Pattern:  utils.Pointer("MMM D, YYYY @ HH:mm:ss.SSS"),
								Timezone: utils.Pointer("America/New_York"),
							},
						},
						"user.is_active": {
							Id: utils.Pointer("boolean"),
						},
						"user.status": {
							Id: utils.Pointer("color"),
							Params: &kbapi.DataViewsFieldformatParams{
								FieldType: utils.Pointer("string"),
								Colors: &[]kbapi.DataViewsFieldformatParamsColor{
									{
										Range:      utils.Pointer("-Infinity:Infinity"),
										Regex:      utils.Pointer("inactive*"),
										Text:       utils.Pointer("#000000"),
										Background: utils.Pointer("#ffffff"),
									},
								},
							},
						},
						"user.message": {
							Id: utils.Pointer("truncate"),
							Params: &kbapi.DataViewsFieldformatParams{
								FieldLength: utils.Pointer(10),
							},
						},
						"host.name": {
							Id: utils.Pointer("string"),
							Params: &kbapi.DataViewsFieldformatParams{
								Transform: utils.Pointer("upper"),
							},
						},
						"response.code": {
							Id: utils.Pointer("static_lookup"),
							Params: &kbapi.DataViewsFieldformatParams{
								LookupEntries: &[]kbapi.DataViewsFieldformatParamsLookup{
									{
										Key:   utils.Pointer("200"),
										Value: utils.Pointer("OK"),
									},
									{
										Key:   utils.Pointer("404"),
										Value: utils.Pointer("Not Found"),
									},
								},
								UnknownKeyValue: utils.Pointer("Unknown"),
							},
						},
						"url.original": {
							Id: utils.Pointer("url"),
							Params: &kbapi.DataViewsFieldformatParams{
								Type:          utils.Pointer("a"),
								UrlTemplate:   utils.Pointer("URL TEMPLATE"),
								LabelTemplate: utils.Pointer("LABEL TEMPLATE"),
							},
						},
						"user.profile_picture": {
							Id: utils.Pointer("url"),
							Params: &kbapi.DataViewsFieldformatParams{
								Type:          utils.Pointer("img"),
								UrlTemplate:   utils.Pointer("URL TEMPLATE"),
								LabelTemplate: utils.Pointer("LABEL TEMPLATE"),
								Width:         utils.Pointer(6),
								Height:        utils.Pointer(4),
							},
						},
						"user.answering_message": {
							Id: utils.Pointer("url"),
							Params: &kbapi.DataViewsFieldformatParams{
								Type:          utils.Pointer("audio"),
								UrlTemplate:   utils.Pointer("URL TEMPLATE"),
								LabelTemplate: utils.Pointer("LABEL TEMPLATE"),
							},
						},
					},
					Name: utils.Pointer("name"),
					RuntimeFieldMap: &map[string]kbapi.DataViewsRuntimefieldmap{
						"runtime_field": {
							Type: "keyword",
							Script: kbapi.DataViewsRuntimefieldmapScript{
								Source: utils.Pointer("emit(\"hello\")"),
							},
						},
					},
					SourceFilters: &[]kbapi.DataViewsSourcefilterItem{
						{Value: "field1"},
						{Value: "field2"},
					},
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         utils.Pointer("title"),
				},
			},
		},
		{
			name: "nil collections",
			model: dataViewModel{
				DataView: utils.ObjectValueFrom(ctx, &innerModel{
					Title:           types.StringValue("title"),
					Name:            types.StringValue("name"),
					ID:              types.StringValue("id"),
					TimeFieldName:   types.StringValue("time_field_name"),
					AllowNoIndex:    types.BoolValue(true),
					SourceFilters:   types.ListNull(types.StringType),
					FieldAttributes: types.MapNull(getFieldAttrElemType()),
					RuntimeFieldMap: types.MapNull(getRuntimeFieldMapElemType()),
					FieldFormats:    types.MapNull(getFieldFormatElemType()),
					Namespaces:      types.ListNull(types.StringType),
				}, getDataViewAttrTypes(), path.Root("data_view"), &diags),
			},
			expectedRequest: kbapi.DataViewsUpdateDataViewRequestObject{
				DataView: kbapi.DataViewsUpdateDataViewRequestObjectInner{
					AllowNoIndex:  utils.Pointer(true),
					Name:          utils.Pointer("name"),
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         utils.Pointer("title"),
				},
			},
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, diags := tt.model.toAPIUpdateModel(ctx)
			require.Equal(t, tt.expectedRequest, request)
			require.Empty(t, diags)
		})
	}
}

func Test_dataViewModel_getViewIDAndSpaceID(t *testing.T) {
	tests := []struct {
		name            string
		model           dataViewModel
		expectedViewID  string
		expectedSpaceID string
	}{
		{
			name: "gets the view and space id from the composite id if set",
			model: dataViewModel{
				ID: types.StringValue("space-id/view-id"),
			},
			expectedViewID:  "view-id",
			expectedSpaceID: "space-id",
		},
		{
			name: "gets the view and space id from the data view if id is not a valid composite id",
			model: dataViewModel{
				ID:      types.StringValue("view-id"),
				SpaceID: types.StringValue("space-id"),
			},
			expectedViewID:  "view-id",
			expectedSpaceID: "space-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewID, spaceID := tt.model.getViewIDAndSpaceID()
			require.Equal(t, tt.expectedViewID, viewID)
			require.Equal(t, tt.expectedSpaceID, spaceID)
		})
	}
}
