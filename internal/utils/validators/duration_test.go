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

package validators

import (
	"errors"
	"reflect"
	"testing"
)

func TestStringIsDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            any
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid duration string",
			i:    "30s",
			k:    "timeout",
		},
		{
			name:       "invalid duration string",
			i:          "30ss",
			k:          "timeout",
			wantErrors: []error{errors.New(`"timeout" contains an invalid duration: time: unknown unit "ss" in duration "30ss"`)},
		},
		{
			name:       "invalid type",
			i:          30,
			k:          "timeout",
			wantErrors: []error{errors.New("expected type of timeout to be string")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsDuration(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsDuration() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(errorsToStrings(gotErrors), errorsToStrings(tt.wantErrors)) {
				t.Errorf("StringIsDuration() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func TestStringIsElasticDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		i            any
		k            string
		wantWarnings []string
		wantErrors   []error
	}{
		{
			name: "valid Elastic duration string",
			i:    "30d",
			k:    "delay",
		},
		{
			name:       "invalid Elastic duration unit",
			i:          "12w",
			k:          "delay",
			wantErrors: []error{errors.New(`"delay" contains an invalid duration: not conforming to Elastic time-units format`)},
		},
		{
			name:       "invalid Elastic duration value",
			i:          ".12s",
			k:          "delay",
			wantErrors: []error{errors.New(`"delay" contains an invalid duration: not conforming to Elastic time-units format`)},
		},
		{
			name:       "invalid data type",
			i:          30,
			k:          "delay",
			wantErrors: []error{errors.New("expected type of delay to be string")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWarnings, gotErrors := StringIsElasticDuration(tt.i, tt.k)
			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("StringIsElasticDuration() gotWarnings = %v, want %v", gotWarnings, tt.wantWarnings)
			}
			if !reflect.DeepEqual(errorsToStrings(gotErrors), errorsToStrings(tt.wantErrors)) {
				t.Errorf("StringIsElasticDuration() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func errorsToStrings(errs []error) []string {
	if errs == nil {
		return nil
	}
	out := make([]string, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			out = append(out, "")
			continue
		}
		out = append(out, err.Error())
	}
	return out
}
