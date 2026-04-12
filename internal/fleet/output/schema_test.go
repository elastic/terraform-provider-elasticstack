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

package output

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaIncludesRemoteElasticsearchTypeAndServiceToken(t *testing.T) {
	t.Parallel()

	s := getSchema()

	typeAttr, ok := s.Attributes["type"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, typeAttr.Validators)

	allowedType := false
	for _, validator := range typeAttr.Validators {
		if strings.Contains(validator.Description(context.Background()), "remote_elasticsearch") {
			allowedType = true
			break
		}
	}
	assert.True(t, allowedType, "expected remote_elasticsearch to be an allowed type")

	serviceTokenAttr, ok := s.Attributes["service_token"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, serviceTokenAttr.Sensitive)
	assert.True(t, serviceTokenAttr.Optional)
	assert.NotEmpty(t, serviceTokenAttr.Validators)
}
