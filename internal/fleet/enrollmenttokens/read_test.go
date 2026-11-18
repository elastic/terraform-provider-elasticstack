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

package enrollmenttokens

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPolicyFilter(t *testing.T) {
	tests := map[string]struct {
		policyID types.String
		want     string
	}{
		"null":        {types.StringNull(), ""},
		"unknown":     {types.StringUnknown(), ""},
		"empty":       {types.StringValue(""), ""},
		"placeholder": {types.StringValue("_"), "_"},
		"real":        {types.StringValue("policy-1"), "policy-1"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, policyFilter(enrollmentTokensModel{PolicyID: tc.policyID}))
		})
	}
}
