package data_view

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/data_views"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/require"
)

func Test_tfModelV0_ToCreateRequest(t *testing.T) {
	tests := []struct {
		name            string
		model           apiModelV0
		expectedRequest data_views.CreateDataViewRequestObject
		expectedDiags   diag.Diagnostics
	}{
		{
			name: "all fields",
			model: apiModelV0{
				SpaceID: "default",
				DataView: apiDataViewV0{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					ID:            "id",
					TimeFieldName: utils.Pointer("time_field_name"),
					SourceFilters: []string{"field1", "field2"},
					FieldAttributes: map[string]apiFieldAttrsV0{
						"field1": {
							CustomLabel: utils.Pointer("custom_label"),
							Count:       utils.Pointer[int64](10),
						},
					},
					RuntimeFieldMap: map[string]apiRuntimeFieldV0{
						"runtime_field": {
							Type:         "keyword",
							ScriptSource: "emit(\"hello\")",
						},
					},
					FieldFormats: map[string]apiFieldFormat{
						"field1": {
							ID: "field1",
							Params: &apiFieldFormatParams{
								Pattern:       "0.00",
								Urltemplate:   "https://test.com/{{value}}",
								Labeltemplate: "{{value}}",
							},
						},
					},
					AllowNoIndex: true,
					Namespaces:   []string{"backend", "o11y"},
				},
			},
			expectedRequest: data_views.CreateDataViewRequestObject{
				Override: utils.Pointer(false),
				DataView: data_views.CreateDataViewRequestObjectDataView{
					AllowNoIndex: utils.Pointer(true),
					FieldAttrs: map[string]interface{}{
						"field1": fieldAttr{
							CustomLabel: utils.Pointer("custom_label"),
							Count:       utils.Pointer[int64](10),
						},
					},
					FieldFormats: map[string]interface{}{
						"field1": apiFieldFormat{
							ID: "field1",
							Params: &apiFieldFormatParams{
								Pattern:       "0.00",
								Urltemplate:   "https://test.com/{{value}}",
								Labeltemplate: "{{value}}",
							},
						},
					},
					Id:         utils.Pointer("id"),
					Name:       utils.Pointer("name"),
					Namespaces: []string{"backend", "o11y", "default"},
					RuntimeFieldMap: map[string]interface{}{
						"runtime_field": runtimeField{
							Type: "keyword",
							Script: runtimeFieldSource{
								Source: "emit(\"hello\")",
							},
						},
					},
					SourceFilters: []data_views.SourcefiltersInner{
						{Value: "field1"},
						{Value: "field2"},
					},
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         "title",
				},
			},
		},
		{
			name: "nil collections",
			model: apiModelV0{
				SpaceID: "default",
				DataView: apiDataViewV0{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					ID:            "id",
					TimeFieldName: utils.Pointer("time_field_name"),
					AllowNoIndex:  true,
				},
			},
			expectedRequest: data_views.CreateDataViewRequestObject{
				Override: utils.Pointer(false),
				DataView: data_views.CreateDataViewRequestObjectDataView{
					AllowNoIndex:  utils.Pointer(true),
					Id:            utils.Pointer("id"),
					Name:          utils.Pointer("name"),
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         "title",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tfModel tfModelV0
			diags := tfsdk.ValueFrom(context.Background(), tt.model, getSchema().Type(), &tfModel)
			require.Nil(t, diags)

			req, diags := tfModel.ToCreateRequest(context.Background())

			require.Equal(t, tt.expectedRequest, req)
			require.Equal(t, tt.expectedDiags, diags)
		})
	}
}

func Test_tfModelV0_ToUpdateRequest(t *testing.T) {
	tests := []struct {
		name            string
		model           apiModelV0
		expectedRequest data_views.UpdateDataViewRequestObject
		expectedDiags   diag.Diagnostics
	}{
		{
			name: "all fields",
			model: apiModelV0{
				DataView: apiDataViewV0{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					ID:            "id",
					TimeFieldName: utils.Pointer("time_field_name"),
					SourceFilters: []string{"field1", "field2"},
					FieldAttributes: map[string]apiFieldAttrsV0{
						"field1": {
							CustomLabel: utils.Pointer("custom_label"),
							Count:       utils.Pointer[int64](10),
						},
					},
					RuntimeFieldMap: map[string]apiRuntimeFieldV0{
						"runtime_field": {
							Type:         "keyword",
							ScriptSource: "emit(\"hello\")",
						},
					},
					FieldFormats: map[string]apiFieldFormat{
						"field1": {
							ID: "field1",
							Params: &apiFieldFormatParams{
								Pattern:       "0.00",
								Urltemplate:   "https://test.com/{{value}}",
								Labeltemplate: "{{value}}",
							},
						},
					},
					AllowNoIndex: true,
					Namespaces:   []string{"default", "o11y"},
				},
			},
			expectedRequest: data_views.UpdateDataViewRequestObject{
				DataView: data_views.UpdateDataViewRequestObjectDataView{
					AllowNoIndex: utils.Pointer(true),
					FieldFormats: map[string]interface{}{
						"field1": apiFieldFormat{
							ID: "field1",
							Params: &apiFieldFormatParams{
								Pattern:       "0.00",
								Urltemplate:   "https://test.com/{{value}}",
								Labeltemplate: "{{value}}",
							},
						},
					},
					Name: utils.Pointer("name"),
					RuntimeFieldMap: map[string]interface{}{
						"runtime_field": runtimeField{
							Type: "keyword",
							Script: runtimeFieldSource{
								Source: "emit(\"hello\")",
							},
						},
					},
					SourceFilters: []data_views.SourcefiltersInner{
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
			model: apiModelV0{
				DataView: apiDataViewV0{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					ID:            "id",
					TimeFieldName: utils.Pointer("time_field_name"),
					AllowNoIndex:  true,
				},
			},
			expectedRequest: data_views.UpdateDataViewRequestObject{
				DataView: data_views.UpdateDataViewRequestObjectDataView{
					AllowNoIndex:  utils.Pointer(true),
					Name:          utils.Pointer("name"),
					TimeFieldName: utils.Pointer("time_field_name"),
					Title:         utils.Pointer("title"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tfModel tfModelV0
			diags := tfsdk.ValueFrom(context.Background(), tt.model, getSchema().Type(), &tfModel)
			require.Nil(t, diags)

			req, diags := tfModel.ToUpdateRequest(context.Background())

			require.Equal(t, tt.expectedRequest, req)
			require.Equal(t, tt.expectedDiags, diags)
		})
	}
}

func Test_tfModelV0_FromResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      data_views.DataViewResponseObject
		existingModel apiModelV0
		expectedModel apiModelV0
		expectedDiags diag.Diagnostics
	}{
		{
			name: "all fields",
			existingModel: apiModelV0{
				ID:      "existing-id",
				SpaceID: "existing-space-id",
				DataView: apiDataViewV0{
					Namespaces: []string{"existing-namespace"},
				},
			},
			response: data_views.DataViewResponseObject{
				DataView: &data_views.DataViewResponseObjectDataView{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					Id:            utils.Pointer("id"),
					TimeFieldName: utils.Pointer("time_field_name"),
					AllowNoIndex:  utils.Pointer(true),
					SourceFilters: []data_views.SourcefiltersInner{
						{Value: "field1"},
						{Value: "field2"},
					},
					FieldAttrs: map[string]interface{}{
						"field1": map[string]interface{}{
							"customLabel": "custom_label",
							"count":       10.0,
						},
					},
					FieldFormats: map[string]interface{}{
						"field1": map[string]interface{}{
							"id": "field1",
						},
					},
					RuntimeFieldMap: map[string]interface{}{
						"runtime_field": map[string]interface{}{
							"type": "keyword",
							"script": map[string]interface{}{
								"source": "emit('hello')",
							},
						},
					},
				},
			},
			expectedModel: apiModelV0{
				ID:      "existing-id",
				SpaceID: "existing-space-id",
				DataView: apiDataViewV0{
					Title:         utils.Pointer("title"),
					Name:          utils.Pointer("name"),
					ID:            "id",
					TimeFieldName: utils.Pointer("time_field_name"),
					SourceFilters: []string{"field1", "field2"},
					FieldAttributes: map[string]apiFieldAttrsV0{
						"field1": {
							CustomLabel: utils.Pointer("custom_label"),
							Count:       utils.Pointer[int64](10),
						},
					},
					RuntimeFieldMap: map[string]apiRuntimeFieldV0{
						"runtime_field": {
							Type:         "keyword",
							ScriptSource: "emit('hello')",
						},
					},
					FieldFormats: map[string]apiFieldFormat{
						"field1": {
							ID: "field1",
						},
					},
					AllowNoIndex: true,
					Namespaces:   []string{"existing-namespace"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tfModel tfModelV0
			diags := tfsdk.ValueFrom(context.Background(), tt.existingModel, getSchema().Type(), &tfModel)
			require.Nil(t, diags)

			finalModel, diags := tfModel.FromResponse(context.Background(), &tt.response)

			require.Equal(t, tt.expectedModel, finalModel)
			require.Equal(t, tt.expectedDiags, diags)
		})
	}
}
