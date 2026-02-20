package typeutils_test

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestExpandStringSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		set  *schema.Set
		want []string
	}{
		{
			name: "returns empty",
			set:  schema.NewSet(schema.HashString, []any{}),
			want: nil,
		},
		{
			name: "converts to string array",
			set:  schema.NewSet(schema.HashString, []any{"a", "b", "c"}),
			want: []string{"c", "b", "a"}, // reordered by hash
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeutils.ExpandStringSet(tt.set); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExpandStringSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
