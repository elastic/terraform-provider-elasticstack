// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package schemautil

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	maps0 "maps"
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
	var j, j2 any
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j, j2), nil
}

func MergeSchemaMaps(maps ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)
	for _, m := range maps {
		maps0.Copy(result, m)
	}
	return result
}

func IsEmpty(v any) bool {
	switch t := v.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		if t == 0 {
			return true
		}
	case string:
		if strings.TrimSpace(t) == "" {
			return true
		}
	case []any:
		if len(t) == 0 {
			return true
		}
	case map[any]any:
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

func ExpandIndividuallyDefinedSettings(ctx context.Context, d *schema.ResourceData, settingsKeys map[string]schema.ValueType) map[string]any {
	settings := make(map[string]any)
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

// MapRef is similar to new(value), in that it takes the reference of
// the given value, however if the value is already nil then it returns
// nil rather than a pointer to nil.
func MapRef[T any, M ~map[string]T](value M) *M {
	if value == nil {
		return nil
	}
	return &value
}

// SliceRef is similar to new(value), in that it takes the reference of
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
	}
	return *value
}

// Itol converts *int to *in64.
func Itol(value *int) *int64 {
	if value == nil {
		return nil
	}
	return new(int64(*value))
}

// Ltoi converts *int64 to *int.
func Ltoi(value *int64) *int {
	if value == nil {
		return nil
	}
	return new(int(*value))
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
