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

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const sslAtLeastOneSubattrSummary = "Empty ssl configuration"

var sslSubattrNames = []string{"certificate_authorities", "certificate", "key"}

// sslAtLeastOneSubattrValidator rejects an ssl object whose sub-attributes
// are all null, so a practitioner cannot send an ssl object the Fleet API
// would persist as a present-but-empty object. Unknown values (the whole
// object or any child) intentionally defer validation: they may resolve to a
// set value later in the apply.
type sslAtLeastOneSubattrValidator struct{}

func (sslAtLeastOneSubattrValidator) Description(_ context.Context) string {
	return "At least one of certificate_authorities, certificate, or key must be set"
}

func (v sslAtLeastOneSubattrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (sslAtLeastOneSubattrValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	for _, name := range sslSubattrNames {
		if attr, ok := attrs[name]; ok && !attr.IsNull() {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		sslAtLeastOneSubattrSummary,
		"At least one of certificate_authorities, certificate, or key must be set.",
	)
}
