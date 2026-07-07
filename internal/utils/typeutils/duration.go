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

package typeutils

import (
	"encoding/json"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ElasticsearchDurationToString converts an estypes.Duration to a Terraform types.String.
// Returns types.StringNull() when d is nil.
func ElasticsearchDurationToString(d estypes.Duration) types.String {
	if d == nil {
		return types.StringNull()
	}
	if s, ok := d.(string); ok {
		return types.StringValue(s)
	}
	b, err := json.Marshal(d)
	if err != nil {
		return types.StringNull()
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return types.StringValue(string(b))
	}
	return types.StringValue(s)
}
