package utils_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type naive struct {
	ID string `json:"id"`
}
type aware struct {
	ID types.String `tfsdk:"id"`
}

var (
	naiveNil   = ([]naive)(nil)
	naiveEmpty = []naive{}
	naiveFull  = []naive{
		{ID: "id1"},
		{ID: "id2"},
		{ID: "id3"},
	}

	awareNil   = ([]aware)(nil)
	awareEmpty = []aware{}
	awareFull  = []aware{
		{ID: types.StringValue("id1")},
		{ID: types.StringValue("id2")},
		{ID: types.StringValue("id3")},
	}

	awareType      = types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}
	awareListUnk   = types.ListUnknown(awareType)
	awareListNil   = types.ListNull(awareType)
	awareListEmpty = types.ListValueMust(awareType, []attr.Value{})
	awareListFull  = types.ListValueMust(awareType, []attr.Value{
		types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id1")}),
		types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id2")}),
		types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id3")}),
	})

	toNaive = func(item aware) naive { return naive{ID: item.ID.ValueString()} }
	toAware = func(item naive) aware { return aware{ID: types.StringValue(item.ID)} }

	stringNil   = ([]string)(nil)
	stringEmpty = []string{}
	stringFull  = []string{"v1", "v2", "v3"}

	stringListUnk   = types.ListUnknown(types.StringType)
	stringListNil   = types.ListNull(types.StringType)
	stringListEmpty = types.ListValueMust(types.StringType, []attr.Value{})
	stringListFull  = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("v1"),
		types.StringValue("v2"),
		types.StringValue("v3"),
	})
)

func TestSliceToListType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []naive
		want  types.List
	}{
		{name: "converts nil", input: naiveNil, want: awareListNil},
		{name: "converts empty", input: naiveEmpty, want: awareListEmpty},
		{name: "converts struct", input: naiveFull, want: awareListFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.SliceToListType(ctx, tt.input, awareType, path.Empty(), diags, toAware)
			if !got.Equal(tt.want) {
				t.Errorf("SliceToListType() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("SlicetoListType() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}

func TestSliceToListType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  types.List
	}{
		{name: "converts nil", input: stringNil, want: stringListNil},
		{name: "converts empty", input: stringEmpty, want: stringListEmpty},
		{name: "converts strings", input: stringFull, want: stringListFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.SliceToListType_String(ctx, tt.input, path.Empty(), diags)
			if !got.Equal(tt.want) {
				t.Errorf("SliceToListType_String() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("SliceToListType_String() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}

func TestListTypeToSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		want  []naive
		input types.List
	}{
		{name: "converts unknown", input: awareListUnk, want: naiveNil},
		{name: "converts nil", input: awareListNil, want: naiveNil},
		{name: "converts empty", input: awareListEmpty, want: naiveEmpty},
		{name: "converts struct", input: awareListFull, want: naiveFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToSlice(ctx, tt.input, path.Empty(), diags, toNaive)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTypeToSlice() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("ListTypeToSlice() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}

func TestListTypeToSlice_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.List
		want  []string
	}{
		{name: "converts unknown", input: stringListUnk, want: stringNil},
		{name: "converts nil", input: stringListNil, want: stringNil},
		{name: "converts empty", input: stringListEmpty, want: stringEmpty},
		{name: "converts strings", input: stringListFull, want: stringFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToSlice_String(ctx, tt.input, path.Empty(), diags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTypeToSlice_String() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("ListTypeToSlice_String() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}

func TestListTypeAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		want  []aware
		input types.List
	}{
		{name: "converts unknown", input: awareListUnk, want: awareNil},
		{name: "converts nil", input: awareListNil, want: awareNil},
		{name: "converts empty", input: awareListEmpty, want: awareEmpty},
		{name: "converts struct", input: awareListFull, want: awareFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeAs[aware](ctx, tt.input, path.Empty(), diags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTypeAs() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("ListTypeAs() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}

func TestTransformSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []naive
		want  []aware
	}{
		{name: "converts nil", input: naiveNil, want: awareNil},
		{name: "converts empty", input: naiveEmpty, want: awareEmpty},
		{name: "converts struct", input: naiveFull, want: awareFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformSlice(tt.input, toAware)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransformSlice() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("TransformSlice() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}
