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
							Id: utils.Pointer("field1"),
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
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := tt.existingModel.populateFromAPI(ctx, &tt.response)

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
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtName("field_formats").AtMapKey("field1").AtName("params"), &diags),
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
							}, getFieldFormatParamsAttrTypes(), path.Root("data_view").AtMapKey("field1").AtName("params"), &diags),
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
