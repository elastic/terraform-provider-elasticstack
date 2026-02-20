package schemautil_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/tfsdkutils"
)

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
		if sup := tfsdkutils.DiffIndexSettingSuppress("", tc.old, tc.new, nil); sup != tc.equal {
			t.Errorf("Failed for test case: %+v", tc)
		}
	}
}
