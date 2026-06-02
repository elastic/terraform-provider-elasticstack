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

package repository

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StrSetting extracts a string setting, returning "" on missing or nil value.
func StrSetting(settings map[string]any, key string) string {
	v, ok := settings[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

// StrSettingNull extracts a string setting, returning a null types.String on missing or nil value.
func StrSettingNull(settings map[string]any, key string) types.String {
	v, ok := settings[key]
	if !ok || v == nil {
		return types.StringNull()
	}
	switch val := v.(type) {
	case string:
		return types.StringValue(val)
	default:
		return types.StringValue(fmt.Sprintf("%v", v))
	}
}

// BoolSettingNull extracts a bool setting, returning a null types.Bool on missing or nil value.
// Returns an error if the value is present but cannot be parsed as bool.
func BoolSettingNull(settings map[string]any, key string) (types.Bool, error) {
	v, ok := settings[key]
	if !ok || v == nil {
		return types.BoolNull(), nil
	}
	switch val := v.(type) {
	case bool:
		return types.BoolValue(val), nil
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return types.BoolNull(), fmt.Errorf(`failed to parse value = "%v" for setting = "%s"`, v, key)
		}
		return types.BoolValue(b), nil
	default:
		return types.BoolNull(), fmt.Errorf(`failed to parse value = "%v" for setting = "%s"`, v, key)
	}
}

// BoolSetting extracts a bool setting, returning fallback on missing, nil, or unparseable value.
func BoolSetting(settings map[string]any, key string, fallback bool) bool {
	b, err := BoolSettingNull(settings, key)
	if err != nil || b.IsNull() {
		return fallback
	}
	return b.ValueBool()
}

// Int64SettingNull extracts an int64 setting, returning a null types.Int64 on missing or nil value.
// Returns an error if the value is present but cannot be parsed as int64.
func Int64SettingNull(settings map[string]any, key string) (types.Int64, error) {
	v, ok := settings[key]
	if !ok || v == nil {
		return types.Int64Null(), nil
	}
	switch val := v.(type) {
	case int:
		return types.Int64Value(int64(val)), nil
	case int64:
		return types.Int64Value(val), nil
	case float64:
		return types.Int64Value(int64(val)), nil
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			f, err2 := strconv.ParseFloat(val, 64)
			if err2 != nil {
				return types.Int64Null(), fmt.Errorf(`failed to parse value = "%v" for setting = "%s"`, v, key)
			}
			return types.Int64Value(int64(f)), nil
		}
		return types.Int64Value(i), nil
	default:
		return types.Int64Null(), fmt.Errorf(`failed to parse value = "%v" for setting = "%s"`, v, key)
	}
}

// Int64Setting extracts an int64 setting, returning fallback on missing, nil, or unparseable value.
func Int64Setting(settings map[string]any, key string, fallback int64) int64 {
	i, err := Int64SettingNull(settings, key)
	if err != nil || i.IsNull() {
		return fallback
	}
	return i.ValueInt64()
}
