package utils_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func TestFlattenMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  map[string]any
		out map[string]any
	}{
		{
			map[string]any{"key1": map[string]any{"key2": 1}},
			map[string]any{"key1.key2": 1},
		},
		{
			map[string]any{"key1": map[string]any{"key2": map[string]any{"key3": 1}}},
			map[string]any{"key1.key2.key3": 1},
		},
		{
			map[string]any{"key1": 1},
			map[string]any{"key1": 1},
		},
		{
			map[string]any{"key1": "test"},
			map[string]any{"key1": "test"},
		},
		{
			map[string]any{"key1": map[string]any{"key2": 1, "key3": "test", "key4": []int{1, 2, 3}}},
			map[string]any{"key1.key2": 1, "key1.key3": "test", "key1.key4": []int{1, 2, 3}},
		},
	}

	for _, tc := range tests {
		res := utils.FlattenMap(tc.in)
		if !utils.MapsEqual(res, tc.out) {
			t.Errorf("Could not properly flatten the map: %+v <> %+v", res, tc.out)
		}
	}
}

func TestDiffIndexTemplateSuppress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		old   string
		new   string
		equal bool
	}{
		{
			`{"key1.key2": 2, "index.key2.key1": "3"}`,
			`{"index": {"key1.key2": "2", "key2.key1": "3"}}`,
			true,
		},
		{
			`{"key1": "2", "key2": "3"}`,
			`{"index": {"key1": "2", "key2": "3"}}`,
			true,
		},
		{
			`{"index":{"key1": "2", "key2": "3"}}`,
			`{"index": {"key1": "2", "key2": "3"}}`,
			true,
		},
		{
			`{"key1": "2", "key2": "3"}`,
			`{"index.key1": "2", "index.key2": "3"}`,
			true,
		},
		{
			`{"key1": 1, "key2": 2}`,
			`{"key1": "2", "index.key2": "3"}`,
			false,
		},
	}

	for _, tc := range tests {
		if sup := utils.DiffIndexSettingSuppress("", tc.old, tc.new, nil); sup != tc.equal {
			t.Errorf("Failed for test case: %+v", tc)
		}
	}
}
