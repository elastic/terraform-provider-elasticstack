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

// durationToMsString formats a time.Duration as a millisecond string (e.g. "5000ms")
// for use with typed API builder methods that accept a string timeout.
func durationToMsString(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds(), 10) + "ms"
}
