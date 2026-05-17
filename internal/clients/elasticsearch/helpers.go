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

package elasticsearch

import (
	"errors"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// IsNotFoundElasticsearchError reports whether err is an Elasticsearch API
// error with HTTP status 404. Use this to treat a missing resource as a
// successful no-op (e.g. idempotent deletes) or as a "not found" signal on
// read operations.
func IsNotFoundElasticsearchError(err error) bool {
	var esErr *types.ElasticsearchError
	return errors.As(err, &esErr) && esErr.Status == 404
}

// durationToMsString formats a time.Duration as a millisecond string (e.g. "5000ms")
// for use with typed API builder methods that accept a string timeout.
func durationToMsString(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds(), 10) + "ms"
}

// NormalizeQueryFilter recursively compacts expanded single-key query values
// produced by the typed client back to their shorthand form.
// For example: {"term":{"field":{"value":"x"}}} → {"term":{"field":"x"}}
func NormalizeQueryFilter(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// If this map has exactly one key "value" with a scalar value, compact it.
		if len(val) == 1 {
			if inner, ok := val["value"]; ok {
				switch inner.(type) {
				case string, float64, bool, int, int64:
					return inner
				}
			}
		}
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = NormalizeQueryFilter(vv)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = NormalizeQueryFilter(vv)
		}
		return out
	default:
		return v
	}
}
