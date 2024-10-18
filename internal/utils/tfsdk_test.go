package utils_test

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

	mapNil   = (map[string]naive)(nil)
	mapEmpty = map[string]naive{}
	mapFull  = map[string]naive{
		"k1": {ID: "id1"},
		"k2": {ID: "id2"},
		"k3": {ID: "id3"},
	}

	normUnk   = jsontypes.NewNormalizedUnknown()
	normNil   = jsontypes.NewNormalizedNull()
	normEmpty = jsontypes.NewNormalizedValue(`{}`)
	normFull  = jsontypes.NewNormalizedValue(`{"k1":{"id":"id1"},"k2":{"id":"id2"},"k3":{"id":"id3"}}`)
)

func TestMapToNormalizedType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]naive
		want  jsontypes.Normalized
	}{
		{name: "converts nil", input: mapNil, want: normNil},
		{name: "converts empty", input: mapEmpty, want: normEmpty},
		{name: "converts struct", input: mapFull, want: normFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.MapToNormalizedType(tt.input, path.Empty(), &diags)
			if !got.Equal(tt.want) {
				t.Errorf("MapToNormalizedType() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("MapToNormalizedType() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}

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
			got := utils.SliceToListType(ctx, tt.input, awareType, path.Empty(), &diags,
				func(item naive, meta utils.ListMeta) aware {
					return aware{ID: types.StringValue(item.ID)}
				},
			)
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
			got := utils.SliceToListType_String(ctx, tt.input, path.Empty(), &diags)
			if !got.Equal(tt.want) {
				t.Errorf("SliceToListType_String() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("SliceToListType_String() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
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
		{name: "converts unknown", input: awareListUnk, want: mapNil},
		{name: "converts nil", input: awareListNil, want: mapNil},
		{name: "converts empty", input: awareListEmpty, want: mapEmpty},
		{name: "converts struct", input: awareListFull, want: mapFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToMap(ctx, tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ListMeta) (string, naive) {
					return "k" + item.ID.ValueString()[2:], toNaive(item)
				})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTypeToMap() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("ListTypeToMap() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
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
		{name: "converts unknown", input: awareListUnk, want: naiveNil},
		{name: "converts nil", input: awareListNil, want: naiveNil},
		{name: "converts empty", input: awareListEmpty, want: naiveEmpty},
		{name: "converts struct", input: awareListFull, want: naiveFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListTypeToSlice(ctx, tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ListMeta) naive {
					return toNaive(item)
				})
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
			got := utils.ListTypeToSlice_String(ctx, tt.input, path.Empty(), &diags)
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
			got := utils.ListTypeAs[aware](ctx, tt.input, path.Empty(), &diags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTypeAs() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("ListTypeAs() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
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
		{name: "converts nil", input: awareNil, want: awareListNil},
		{name: "converts empty", input: awareEmpty, want: awareListEmpty},
		{name: "converts struct", input: awareFull, want: awareListFull},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.ListValueFrom(ctx, tt.input, awareType, path.Empty(), &diags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListValueFrom() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("ListTypeAs() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
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
		{name: "converts unknown", input: normUnk, want: mapNil},
		{name: "converts nil", input: normNil, want: mapNil},
		{name: "converts empty", input: normEmpty, want: mapEmpty},
		{name: "converts struct", input: normFull, want: mapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.NormalizedTypeToMap[naive](tt.input, path.Empty(), &diags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapToNormalizedType() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("MapToNormalizedType() diagnostic: %s: %s", d.Summary(), d.Detail())
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
			got := utils.TransformSlice(tt.input, path.Empty(), &diags,
				func(item naive, meta utils.ListMeta) aware {
					return toAware(item)
				})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransformSlice() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("TransformSlice() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
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
		{name: "converts nil", input: awareNil, want: mapNil},
		{name: "converts empty", input: awareEmpty, want: mapEmpty},
		{name: "converts struct", input: awareFull, want: mapFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformSliceToMap(tt.input, path.Empty(), &diags,
				func(item aware, meta utils.ListMeta) (string, naive) {
					return "k" + item.ID.ValueString()[2:], toNaive(item)
				})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransformSliceToMap() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("TransformSliceToMap() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
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
		{name: "converts nil", input: mapNil, want: naiveNil},
		{name: "converts empty", input: mapEmpty, want: naiveEmpty},
		{name: "converts struct", input: mapFull, want: naiveFull},
	}

	sortFn := func(s []naive) func(i, j int) bool {
		return func(i, j int) bool {
			return s[i].ID < s[j].ID
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.TransformMapToSlice(tt.input, path.Empty(), &diags,
				func(item naive, meta utils.MapMeta) naive {
					return item
				})

			sort.Slice(got, sortFn(got))
			sort.Slice(tt.want, sortFn(tt.want))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransformMapToSlice() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("TransformMapToSlice() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}
