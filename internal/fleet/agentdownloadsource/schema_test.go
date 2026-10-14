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

package agentdownloadsource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSchemaSsl(t *testing.T) {
	ctx := context.Background()

	sslAttr, ok := getSchema(ctx).Attributes["ssl"].(schema.SingleNestedAttribute)
	require.True(t, ok, "ssl attribute must exist as a single nested attribute")
	require.True(t, sslAttr.Optional, "ssl must be Optional")
	require.Equal(t, []validator.Object{sslAtLeastOneSubattrValidator{}}, sslAttr.Validators)

	cas, ok := sslAttr.Attributes["certificate_authorities"].(schema.ListAttribute)
	require.True(t, ok)
	require.Equal(t, types.StringType, cas.ElementType)
	require.True(t, cas.Optional)
	require.Equal(t, []validator.List{
		listvalidator.SizeAtLeast(1),
		listvalidator.SizeAtMost(10),
		listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
	}, cas.Validators)

	cert, ok := sslAttr.Attributes["certificate"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, cert.Optional)
	require.Equal(t, []validator.String{stringvalidator.LengthAtLeast(1)}, cert.Validators)

	key, ok := sslAttr.Attributes["key"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, key.Optional)
	require.True(t, key.Sensitive)
	require.Equal(t, []validator.String{stringvalidator.LengthAtLeast(1)}, key.Validators)

	_, hasVerificationMode := sslAttr.Attributes["verification_mode"]
	require.False(t, hasVerificationMode, "ssl must not expose verification_mode")
}
