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

package customtypes

import (
	"encoding/json"
	"errors"
	"strings"
)

// ErrInvalidStringSliceOrCSV is returned when UnmarshalJSON receives an unexpected type.
var ErrInvalidStringSliceOrCSV = errors.New("expected array of strings, or a csv string")

// StringSliceOrCSV deserialises a JSON field that can be either an array ["a","b"]
// or a comma-separated string "a,b".
type StringSliceOrCSV []string

func (i *StringSliceOrCSV) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// First try to parse as an array
	var sliceResult []string
	if err := json.Unmarshal(data, &sliceResult); err == nil {
		*i = StringSliceOrCSV(sliceResult)
		return nil
	}

	var stringResult string
	if err := json.Unmarshal(data, &stringResult); err == nil {
		*i = StringSliceOrCSV(strings.Split(stringResult, ","))
		return nil
	}

	return ErrInvalidStringSliceOrCSV
}
