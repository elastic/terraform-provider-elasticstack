package utils_test

import (
	"context"
	"sort"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

type naive struct {
	ID string `json:"id"`
}

type aware struct {
	ID types.String `tfsdk:"id"`
}

var (
	naiveSliceNil   = ([]naive)(nil)
	naiveSliceEmpty = []naive{}
	naiveSliceFull  = []naive{
		{ID: "id1"},
		{ID: "id2"},
		{ID: "id3"},
	}

	naiveMapNil   = (map[string]naive)(nil)
	naiveMapEmpty = map[string]naive{}
	naiveMapFull  = map[string]naive{
		"k1": {ID: "id1"},
		"k2": {ID: "id2"},
		"k3": {ID: "id3"},
	}

	naiveStructNil  = (*naive)(nil)
	naiveStructFull = &naive{ID: "val"}

	awareSliceNil   = ([]aware)(nil)
	awareSliceEmpty = []aware{}
	awareSliceFull  = []aware{
		{ID: types.StringValue("id1")},
		{ID: types.StringValue("id2")},
		{ID: types.StringValue("id3")},
	}

	awareMapNil   = (map[string]aware)(nil)
	awareMapEmpty = map[string]aware{}
	awareMapFull  = map[string]aware{
		"k1": {ID: types.StringValue("id1")},
		"k2": {ID: types.StringValue("id2")},
		"k3": {ID: types.StringValue("id3")},
	}

	awareStructNil  = (*aware)(nil)
	awareStructFull = &aware{ID: types.StringValue("val")}

	awareType      = types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}
	awareListUnk   = types.ListUnknown(awareType)
	awareListNil   = types.ListNull(awareType)
	awareListEmpty = types.ListValueMust(awareType, []attr.Value{})
	awareListFull  = types.ListValueMust(awareType, []attr.Value{
		types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id1")}),
		types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id2")}),
		types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id3")}),
	})

	awareMapTypeUnk   = types.MapUnknown(awareType)
	awareMapTypeNil   = types.MapNull(awareType)
	awareMapTypeEmpty = types.MapValueMust(awareType, map[string]attr.Value{})
	awareMapTypeFull  = types.MapValueMust(awareType, map[string]attr.Value{
		"k1": types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id1")}),
		"k2": types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id2")}),
		"k3": types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id3")}),
	})

	awareObjectUnk  = types.ObjectUnknown(awareType.AttrTypes)
	awareObjectNil  = types.ObjectNull(awareType.AttrTypes)
	awareObjectFull = types.ObjectValueMust(awareType.AttrTypes, map[string]attr.Value{"id": types.StringValue("val")})

	toNaive = func(item aware) naive { return naive{ID: item.ID.ValueString()} }
	toAware = func(item naive) aware { return aware{ID: types.StringValue(item.ID)} }

	stringSliceNil   = ([]string)(nil)
	stringSliceEmpty = []string{}
	stringSliceFull  = []string{"v1", "v2", "v3"}

	stringListUnk   = types.ListUnknown(types.StringType)
	stringListNil   = types.ListNull(types.StringType)
	stringListEmpty = types.ListValueMust(types.StringType, []attr.Value{})
	stringListFull  = types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("v1"),
		types.StringValue("v2"),
		types.StringValue("v3"),
	})

	normUnk   = jsontypes.NewNormalizedUnknown()
	normNil   = jsontypes.NewNormalizedNull()
	normEmpty = jsontypes.NewNormalizedValue(`{}`)
	normFull  = jsontypes.NewNormalizedValue(`{"k1":{"id":"id1"},"k2":{"id":"id2"},"k3":{"id":"id3"}}`)
)

// Primitives

func TestValueStringPointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.String
		want  *string
	}{
		{name: "converts unknown", input: types.StringUnknown(), want: nil},
		{name: "converts nil", input: types.StringNull(), want: nil},
		{name: "converts empty", input: types.StringValue(""), want: utils.Pointer("")},
		{name: "converts value", input: types.StringValue("value"), want: utils.Pointer("value")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ValueStringPointer(tt.input)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

// Maps

func TestMapToNormalizedType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]naive
		want  jsontypes.Normalized
	}{
		{name: "converts nil", input: naiveMapNil, want: normNil},
		{name: "converts empty", input: naiveMapEmpty, want: normEmpty},
		{name: "converts struct", input: naiveMapFull, want: normFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.MapToNormalizedType(tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestNormalizedTypeToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input jsontypes.Normalized
		want  map[string]naive
	}{
		{name: "converts unknown", input: normUnk, want: naiveMapNil},
		{name: "converts nil", input: normNil, want: naiveMapNil},
		{name: "converts empty", input: normEmpty, want: naiveMapEmpty},
		{name: "converts struct", input: normFull, want: naiveMapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.NormalizedTypeToMap[naive](tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestMapToMapType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]naive
		want  types.Map
	}{
		{name: "converts nil", input: naiveMapNil, want: awareMapTypeNil},
		{name: "converts empty", input: naiveMapEmpty, want: awareMapTypeEmpty},
		{name: "converts struct", input: naiveMapFull, want: awareMapTypeFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.MapToMapType(context.Background(), tt.input, awareType, path.Empty(), &diags,
				func(item naive, meta utils.MapMeta) aware {
					return aware{ID: types.StringValue(item.ID)}
				},
			)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestMapTypeToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.Map
		want  map[string]naive
	}{
		{name: "converts unknown", input: awareMapTypeUnk, want: naiveMapNil},
		{name: "converts nil", input: awareMapTypeNil, want: naiveMapNil},
		{name: "converts empty", input: awareMapTypeEmpty, want: naiveMapEmpty},
		{name: "converts struct", input: awareMapTypeFull, want: naiveMapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.MapTypeToMap(context.Background(), tt.input, path.Empty(), &diags,
				func(item aware, meta utils.MapMeta) naive {
					return toNaive(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}

}

func TestMapTypeAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.Map
		want  map[string]aware
	}{
		{name: "converts unknown", input: awareMapTypeUnk, want: awareMapNil},
		{name: "converts nil", input: awareMapTypeNil, want: awareMapNil},
		{name: "converts empty", input: awareMapTypeEmpty, want: awareMapEmpty},
		{name: "converts struct", input: awareMapTypeFull, want: awareMapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.MapTypeAs[aware](context.Background(), tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestMapValueFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]aware
		want  types.Map
	}{
		{name: "converts nil", input: awareMapNil, want: awareMapTypeNil},
		{name: "converts empty", input: awareMapEmpty, want: awareMapTypeEmpty},
		{name: "converts struct", input: awareMapFull, want: awareMapTypeFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.MapValueFrom(context.Background(), tt.input, awareType, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

// Lists

func TestSliceToListType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []naive
		want  types.List
	}{
		{name: "converts nil", input: naiveSliceNil, want: awareListNil},
		{name: "converts empty", input: naiveSliceEmpty, want: awareListEmpty},
		{name: "converts struct", input: naiveSliceFull, want: awareListFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.SliceToListType(context.Background(), tt.input, awareType, path.Empty(), &diags,
				func(item naive, meta utils.ListMeta) aware {
					return aware{ID: types.StringValue(item.ID)}
				},
			)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
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
		{name: "converts nil", input: stringSliceNil, want: stringListNil},
		{name: "converts empty", input: stringSliceEmpty, want: stringListEmpty},
		{name: "converts strings", input: stringSliceFull, want: stringListFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.SliceToListType_String(context.Background(), tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestListTypeToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.List
		want  map[string]naive
	}{
		{name: "converts unknown", input: awareListUnk, want: naiveMapNil},
		{name: "converts nil", input: awareListNil, want: naiveMapNil},
		{name: "converts empty", input: awareListEmpty, want: naiveMapEmpty},
		{name: "converts struct", input: awareListFull, want: naiveMapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToMap(context.Background(), tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ListMeta) (string, naive) {
					return "k" + item.ID.ValueString()[2:], toNaive(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestListTypeToSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.List
		want  []naive
	}{
		{name: "converts unknown", input: awareListUnk, want: naiveSliceNil},
		{name: "converts nil", input: awareListNil, want: naiveSliceNil},
		{name: "converts empty", input: awareListEmpty, want: naiveSliceEmpty},
		{name: "converts struct", input: awareListFull, want: naiveSliceFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToSlice(context.Background(), tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ListMeta) naive {
					return toNaive(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
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
		{name: "converts unknown", input: stringListUnk, want: stringSliceNil},
		{name: "converts nil", input: stringListNil, want: stringSliceNil},
		{name: "converts empty", input: stringListEmpty, want: stringSliceEmpty},
		{name: "converts strings", input: stringListFull, want: stringSliceFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToSlice_String(context.Background(), tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestListTypeAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.List
		want  []aware
	}{
		{name: "converts unknown", input: awareListUnk, want: awareSliceNil},
		{name: "converts nil", input: awareListNil, want: awareSliceNil},
		{name: "converts empty", input: awareListEmpty, want: awareSliceEmpty},
		{name: "converts struct", input: awareListFull, want: awareSliceFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeAs[aware](context.Background(), tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestListValueFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []aware
		want  types.List
	}{
		{name: "converts nil", input: awareSliceNil, want: awareListNil},
		{name: "converts empty", input: awareSliceEmpty, want: awareListEmpty},
		{name: "converts struct", input: awareSliceFull, want: awareListFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListValueFrom(context.Background(), tt.input, awareType, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

// Objects

func TestStructToObjectType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *naive
		want  types.Object
	}{
		{name: "converts nil", input: naiveStructNil, want: awareObjectNil},
		{name: "converts struct", input: naiveStructFull, want: awareObjectFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.StructToObjectType(context.Background(), tt.input, awareType.AttrTypes, path.Empty(), &diags,
				func(item naive, meta utils.ObjectMeta) aware {
					return aware{ID: types.StringValue(item.ID)}
				},
			)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestObjectTypeToStruct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.Object
		want  *naive
	}{
		{name: "converts unknown", input: awareObjectUnk, want: naiveStructNil},
		{name: "converts nil", input: awareObjectNil, want: naiveStructNil},
		{name: "converts struct", input: awareObjectFull, want: naiveStructFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ObjectTypeToStruct(context.Background(), tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ObjectMeta) naive {
					return toNaive(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestObjectTypeAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input types.Object
		want  *aware
	}{
		{name: "converts unknown", input: awareObjectUnk, want: awareStructNil},
		{name: "converts nil", input: awareObjectNil, want: awareStructNil},
		{name: "converts struct", input: awareObjectFull, want: awareStructFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ObjectTypeAs[aware](context.Background(), tt.input, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestObjectValueFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *aware
		want  types.Object
	}{
		{name: "converts nil", input: awareStructNil, want: awareObjectNil},
		{name: "converts struct", input: awareStructFull, want: awareObjectFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ObjectValueFrom(context.Background(), tt.input, awareType.AttrTypes, path.Empty(), &diags)
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

// Transforms

func TestTransformObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *naive
		want  *aware
	}{
		{name: "converts nil", input: (*naive)(nil), want: (*aware)(nil)},
		{name: "converts struct", input: &naiveSliceFull[0], want: &awareSliceFull[0]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformObject(context.Background(), tt.input, path.Empty(), &diags,
				func(item naive, meta utils.ObjectMeta) aware {
					return toAware(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestTransformMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]naive
		want  map[string]aware
	}{
		{name: "converts nil", input: naiveMapNil, want: awareMapNil},
		{name: "converts empty", input: naiveMapEmpty, want: awareMapEmpty},
		{name: "converts struct", input: naiveMapFull, want: awareMapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformMap(context.Background(), tt.input, path.Empty(), &diags,
				func(item naive, meta utils.MapMeta) aware {
					return toAware(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
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
		{name: "converts nil", input: naiveSliceNil, want: awareSliceNil},
		{name: "converts empty", input: naiveSliceEmpty, want: awareSliceEmpty},
		{name: "converts struct", input: naiveSliceFull, want: awareSliceFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformSlice(context.Background(), tt.input, path.Empty(), &diags,
				func(item naive, meta utils.ListMeta) aware {
					return toAware(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestTransformSliceToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []aware
		want  map[string]naive
	}{
		{name: "converts nil", input: awareSliceNil, want: naiveMapNil},
		{name: "converts empty", input: awareSliceEmpty, want: naiveMapEmpty},
		{name: "converts struct", input: awareSliceFull, want: naiveMapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformSliceToMap(context.Background(), tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ListMeta) (string, naive) {
					return "k" + item.ID.ValueString()[2:], toNaive(item)
				})
			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}

func TestTransformMapToSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]naive
		want  []naive
	}{
		{name: "converts nil", input: naiveMapNil, want: naiveSliceNil},
		{name: "converts empty", input: naiveMapEmpty, want: naiveSliceEmpty},
		{name: "converts struct", input: naiveMapFull, want: naiveSliceFull},
	}

	sortFn := func(s []naive) func(i, j int) bool {
		return func(i, j int) bool {
			return s[i].ID < s[j].ID
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformMapToSlice(context.Background(), tt.input, path.Empty(), &diags,
				func(item naive, meta utils.MapMeta) naive {
					return item
				})

			sort.Slice(got, sortFn(got))
			sort.Slice(tt.want, sortFn(tt.want))

			require.Equal(t, tt.want, got)
			require.Empty(t, diags)
		})
	}
}
