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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	return strings.Replace(settingKey, ".", "_", -1)
}

func Pointer[T any](value T) *T {
	return &value
}

// MapRef is similar to Pointer, in that it takes the reference of
// the given value, however if the value is already nil then it returns
// nil rather than a pointer to nil.
func MapRef[T any](value map[string]T) *map[string]T {
	if value == nil {
		return nil
	}
	return &value
}

// SliceRef is similar to Pointer, in that it takes the reference of
// the given value, however if the value is already nil then it returns
// nil rather than a pointer to nil.
func SliceRef[T any](value []T) *[]T {
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

func FlipMap[K comparable, V comparable](m map[K]V) map[V]K {
	inv := make(map[V]K)
	for k, v := range m {
		inv[v] = k
	}
	return inv
}

func SdkDiagsAsError(diags sdkdiag.Diagnostics) error {
	for _, diag := range diags {
		if diag.Severity == sdkdiag.Error {
			return fmt.Errorf("%s: %s", diag.Summary, diag.Detail)
		}
	}
	return nil
}

func FwDiagsAsError(diags fwdiag.Diagnostics) error {
	for _, diag := range diags {
		if diag.Severity() == fwdiag.SeverityError {
			return fmt.Errorf("%s: %s", diag.Summary(), diag.Detail())
		}
	}
	return nil
}

// ConvertToAttrDiags wraps an existing collection of diagnostics with an attribute path.
func ConvertToAttrDiags(diags fwdiag.Diagnostics, path path.Path) fwdiag.Diagnostics {
	var nd fwdiag.Diagnostics
	for _, d := range diags {
		if d.Severity() == fwdiag.SeverityError {
			nd.AddAttributeError(path, d.Summary(), d.Detail())
		} else if d.Severity() == diag.SeverityWarning {
			nd.AddAttributeWarning(path, d.Summary(), d.Detail())
		} else {
			nd.Append(d)
		}
	}
	return nd
}
