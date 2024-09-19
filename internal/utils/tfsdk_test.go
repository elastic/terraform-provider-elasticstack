package utils_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSliceToListType(t *testing.T) {
	t.Parallel()

	type Type1 struct {
		ID string `json:"id"`
	}
	type Type2 struct {
		ID types.String `tfsdk:"id"`
	}
	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id": types.StringType,
		},
	}
	t1_t2 := func(item any) any {
		i := item.(Type1)
		return Type2{
			ID: types.StringValue(i.ID),
		}
	}
	toString := func(item any) any {
		return types.StringValue(item.(string))
	}

	tests := []struct {
		name     string
		input    []any
		want     types.List
		elemType attr.Type
		iter     func(any) any
	}{
		{
			name:     "converts nil",
			input:    nil,
			want:     types.ListNull(elemType),
			elemType: elemType,
			iter:     t1_t2,
		},
		{
			name:     "converts empty",
			input:    []any{},
			want:     types.ListValueMust(elemType, []attr.Value{}),
			elemType: elemType,
			iter:     t1_t2,
		},
		{
			name: "converts struct",
			input: []any{
				Type1{ID: "id1"},
				Type1{ID: "id2"},
				Type1{ID: "id3"},
			},
			want: types.ListValueMust(elemType, []attr.Value{
				types.ObjectValueMust(elemType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id1")}),
				types.ObjectValueMust(elemType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id2")}),
				types.ObjectValueMust(elemType.AttrTypes, map[string]attr.Value{"id": types.StringValue("id3")}),
			}),
			elemType: elemType,
			iter:     t1_t2,
		},
		{
			name:  "convert strings",
			input: []any{"val1", "val2", "val3"},
			want: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("val1"),
				types.StringValue("val2"),
				types.StringValue("val3"),
			}),
			elemType: types.StringType,
			iter:     toString,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := utils.SliceToListType(ctx, tt.input, tt.elemType, path.Empty(), diags, tt.iter)
			if !got.Equal(tt.want) {
				t.Errorf("SliceToListType() = %v, want %v", got, tt.want)
			}
			for _, d := range diags.Errors() {
				t.Errorf("SlicetoListType() diagnostic: %s: %s", d.Summary(), d.Detail())
			}
		})
	}
}
