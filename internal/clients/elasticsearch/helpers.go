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
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// jsonNullLiteral is the JSON `null` token. Compared against marshaled
// json.RawMessage / []byte values to treat an explicit JSON null the same as
// an absent field.
const jsonNullLiteral = "null"

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
//
// The typed go-elasticsearch/v8 client decodes most API error bodies into
// *types.ElasticsearchError (see generated *Do methods). If a specific endpoint
// ever returns a different error type, extend this helper and update any
// live-stack regression test that asserts the shape for that call path.
func IsNotFoundElasticsearchError(err error) bool {
	if err == nil {
		return false
	}
	var esErr *types.ElasticsearchError
	if !errors.As(err, &esErr) || esErr == nil {
		return false
	}
	return esErr.Status == 404
}

// CallOrNotFound runs fn and applies the standard 404-as-not-found convention
// for typed-client calls made via .Do(ctx): a 404 is swallowed into a zero
// value with no error (signalling "does not exist" to the caller), and any
// other error is wrapped into framework diagnostics using summary.
func CallOrNotFound[T any](fn func() (T, error), summary string) (T, fwdiags.Diagnostics) {
	result, err := fn()
	if err != nil {
		var zero T
		return zero, DeleteWithNotFoundAsSuccess(err, summary)
	}
	return result, nil
}

// DeleteWithNotFoundAsSuccess converts err from a typed-client call made via
// .Do(ctx) into framework diagnostics, treating an Elasticsearch 404 as a
// successful no-op (the resource is already gone) rather than an error. Any
// other error is wrapped into a single diagnostic using summary (e.g.
// "Unable to delete a role").
func DeleteWithNotFoundAsSuccess(err error, summary string) fwdiags.Diagnostics {
	if err == nil || IsNotFoundElasticsearchError(err) {
		return nil
	}
	return diagutil.ErrDiag(summary, err)
}
