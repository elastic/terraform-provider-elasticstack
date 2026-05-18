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
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

// DateMathIndexNameRe matches plain Elasticsearch date math index name expressions.
// The pattern enforces:
//   - opening `<`
//   - a static prefix that starts with a valid non-start character (not -, _, +) and
//     uses only the same character set allowed in ordinary static index names
//   - at least one `{…}` section (the date math expression itself)
//   - a closing `>` immediately after the last `}`
//
// This keeps the two validation paths (static vs date-math) consistent and avoids
// accepting expressions that would be rejected as static names.
var DateMathIndexNameRe = regexp.MustCompile(`^<[^-_+][a-z0-9!$%&'()+.;=@[\]^{}~_-]*\{[^<>]+\}>$`)

// encodeDateMathIndexName URI-encodes a plain date math index name for use in an API
// request path.  Characters inside the expression that have special meaning in a URL
// path are percent-encoded so the Go HTTP client does not rewrite them.
func encodeDateMathIndexName(name string) string {
	// url.PathEscape does not encode '/' by default; we need '/' encoded too
	// so the Go HTTP client does not split the path at that point.
	return strings.ReplaceAll(url.PathEscape(name), "/", "%2F")
}

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

// formatDuration converts a time.Duration to an Elasticsearch timeout string.
// Sub-millisecond values are expressed in nanoseconds (e.g. "500nanos"); all
// other values are expressed in milliseconds (e.g. "5000ms"), matching the
// legacy esapi behavior. Use durationToMsString when sub-ms precision is not
// needed.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return strconv.FormatInt(int64(d), 10) + "nanos"
	}
	return strconv.FormatInt(int64(d)/int64(time.Millisecond), 10) + "ms"
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
