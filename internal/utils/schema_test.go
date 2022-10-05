package utils

import (
	"reflect"
	"testing"

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
			set: schema.NewSet(func(i interface{}) int {
				return schema.HashString(i)
			}, []interface{}{}),
			want: nil,
		},
		{
			name: "converts to string array",
			set: schema.NewSet(func(i interface{}) int {
				return schema.HashString(i)
			}, []interface{}{"a", "b", "c"}),
			want: []string{"c", "b", "a"}, // reordered by hash
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandStringSet(tt.set); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExpandStringSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
