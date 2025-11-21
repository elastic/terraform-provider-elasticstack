package streams

import (
	"encoding/json"
	"reflect"
	"testing"
)

func mustDecodeJSON(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\ninput: %s", err, string(b))
	}
	return v
}

func TestMarshalCondition_SimpleAndOr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cond Condition
		want string
	}{
		{
			name: "single leaf",
			cond: FieldComparison{
				Field: "host.name",
				Op:    "eq",
				Value: "web-01",
			},
			want: `{
				"field": "host.name",
				"op":    "eq",
				"value": "web-01"
			}`,
		},
		{
			name: "and of two leaves",
			cond: And{
				Children: []Condition{
					FieldComparison{Field: "host.name", Op: "eq", Value: "web-01"},
					FieldComparison{Field: "status", Op: "eq", Value: "ok"},
				},
			},
			want: `{
				"and": [
					{"field": "host.name", "op": "eq", "value": "web-01"},
					{"field": "status",    "op": "eq", "value": "ok"}
				]
			}`,
		},
		{
			name: "nested and/or tree",
			cond: And{
				Children: []Condition{
					FieldComparison{Field: "env", Op: "eq", Value: "prod"},
					Or{
						Children: []Condition{
							FieldComparison{Field: "service", Op: "eq", Value: "api"},
							FieldComparison{Field: "service", Op: "eq", Value: "frontend"},
						},
					},
				},
			},
			want: `{
				"and": [
					{"field": "env", "op": "eq", "value": "prod"},
					{"or": [
						{"field": "service", "op": "eq", "value": "api"},
						{"field": "service", "op": "eq", "value": "frontend"}
					]}
				]
			}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotBytes, err := MarshalCondition(tt.cond)
			if err != nil {
				t.Fatalf("MarshalCondition() error = %v", err)
			}

			got := mustDecodeJSON(t, gotBytes)
			want := mustDecodeJSON(t, []byte(tt.want))

			if !reflect.DeepEqual(got, want) {
				gb, _ := json.MarshalIndent(got, "", "  ")
				wb, _ := json.MarshalIndent(want, "", "  ")
				t.Fatalf("mismatch.\nGot:\n%s\nWant:\n%s", string(gb), string(wb))
			}
		})
	}
}
