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
	"fmt"
	"regexp"
	"time"
)

// StringIsDuration is a SchemaValidateFunc which tests to make sure the supplied string is valid duration.
func StringIsDuration(i any, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}

	if _, err := time.ParseDuration(v); err != nil {
		return nil, []error{fmt.Errorf("%q contains an invalid duration: %w", k, err)}
	}

	return nil, nil
}

// StringIsElasticDuration is a SchemaValidateFunc which tests to make sure the supplied string is valid duration using Elastic time units:
// d, h, m, s, ms, micros, nanos. (see the [API conventions time units documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/api-conventions.html#time-units) for more details)
func StringIsElasticDuration(i any, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("%q contains an invalid duration: [empty]", k)}
	}

	r := regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?(?:d|h|m|s|ms|micros|nanos)$`)

	if !r.MatchString(v) {
		return nil, []error{fmt.Errorf("%q contains an invalid duration: not conforming to Elastic time-units format", k)}
	}

	return nil, nil
}
