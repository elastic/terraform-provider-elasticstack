package utils

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Compares the JSON in two byte slices
func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return MapsEqual(j, j2), nil
}

func MapsEqual(m1, m2 interface{}) bool {
	return reflect.DeepEqual(m2, m1)
}

// Flattens the multilevel map, and concatenates keys together with dot "."
// # Examples
// map of form:
//
//	map := map[string]interface{}{
//	        "index": map[string]interface{}{
//	                "key": 1
//	        }
//	}
//
// becomes:
//
//	map := map[string]interface{}{
//	        "index.key": 1
//	}
func FlattenMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	var flattener func(string, map[string]interface{}, map[string]interface{})
	flattener = func(k string, src, dst map[string]interface{}) {
		if len(k) > 0 {
			k += "."
		}
		for key, v := range src {
			switch inner := v.(type) {
			case map[string]interface{}:
				flattener(k+key, inner, dst)
			default:
				dst[k+key] = v
			}
		}
	}
	flattener("", m, out)
	return out
}

func MergeSchemaMaps(maps ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func IsEmpty(v interface{}) bool {
	switch t := v.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		if t == 0 {
			return true
		}
	case string:
		if strings.TrimSpace(t) == "" {
			return true
		}
	case []interface{}:
		if len(t) == 0 {
			return true
		}
	case map[interface{}]interface{}:
		if len(t) == 0 {
			return true
		}
	case nil:
		return true
	}
	return false
}

const connectionKeyName = "elasticsearch_connection"

// Returns the common connection schema for all the Elasticsearch resources,
// which defines the fields which can be used to configure the API access
func AddConnectionSchema(providedSchema map[string]*schema.Schema) {
	providedSchema[connectionKeyName] = providerSchema.GetEsConnectionSchema(connectionKeyName, false)
}

func StringToHash(s string) (*string, error) {
	h := sha1.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		return nil, err
	}
	bs := h.Sum(nil)
	hash := fmt.Sprintf("%x", bs)
	return &hash, nil
}

func FormatStrictDateTime(t time.Time) string {
	strictDateTime := t.Format("2006-01-02T15:04:05.000Z")
	return strictDateTime
}

func ExpandIndividuallyDefinedSettings(ctx context.Context, d *schema.ResourceData, settingsKeys map[string]schema.ValueType) map[string]interface{} {
	settings := make(map[string]interface{})
	for key := range settingsKeys {
		tfFieldKey := ConvertSettingsKeyToTFFieldKey(key)
		if raw, ok := d.GetOk(tfFieldKey); ok {
			switch field := raw.(type) {
			case *schema.Set:
				settings[key] = field.List()
			default:
				settings[key] = raw
			}
			tflog.Trace(ctx, fmt.Sprintf("expandIndividuallyDefinedSettings: settingsKey:%+v tfFieldKey:%+v value:%+v, %+v", key, tfFieldKey, raw, settings))
		}
	}
	return settings
}

func ConvertSettingsKeyToTFFieldKey(settingKey string) string {
	return strings.ReplaceAll(settingKey, ".", "_")
}

func Pointer[T any](value T) *T {
	return &value
}

// MapRef is similar to Pointer, in that it takes the reference of
// the given value, however if the value is already nil then it returns
// nil rather than a pointer to nil.
func MapRef[T any, M ~map[string]T](value M) *M {
	if value == nil {
		return nil
	}
	return &value
}

// SliceRef is similar to Pointer, in that it takes the reference of
// the given value, however if the value is already nil then it returns
// nil rather than a pointer to nil.
func SliceRef[T any, S ~[]T](value S) *S {
	if value == nil {
		return nil
	}
	return &value
}

// Deref returns the value referenced by the given pointer. If the value is
// nil, a zero value is returned.
func Deref[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	} else {
		return *value
	}
}

// Itol converts *int to *in64.
func Itol(value *int) *int64 {
	if value == nil {
		return nil
	}
	return Pointer(int64(*value))
}

// Ltoi converts *int64 to *int.
func Ltoi(value *int64) *int {
	if value == nil {
		return nil
	}
	return Pointer(int(*value))
}

func FlipMap[K comparable, V comparable](m map[K]V) map[V]K {
	inv := make(map[V]K)
	for k, v := range m {
		inv[v] = k
	}
	return inv
}

func DefaultIfNil[T any](value *T) T {
	var result T

	if value != nil {
		result = *value
	}

	return result
}

// Returns an empty slice if s is a slice represented by nil (no backing array).
// Guarantees that json.Marshal and terraform parameters will not treat the
// empty slice as null.
func NonNilSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}

	return s
}

// TimeToStringValue formats a time.Time to ISO 8601 format and returns a types.StringValue.
// This is a convenience function that combines FormatStrictDateTime and types.StringValue.
func TimeToStringValue(t time.Time) types.String {
	return types.StringValue(FormatStrictDateTime(t))
}
