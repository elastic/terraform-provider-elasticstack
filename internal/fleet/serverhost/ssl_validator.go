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

package serverhost

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const sslAtLeastOneSubattrSummary = "Empty ssl configuration"

const sslAtLeastOneSubattrMessage = "At least one of certificate, key, certificate_authorities, or client_auth must be set."

var sslSubattrNames = []string{sslCertificateAttr, sslKeyAttr, sslCertificateAuthoritiesAttr, sslClientAuthAttr}

type sslAtLeastOneSubattrValidator struct{}

func (sslAtLeastOneSubattrValidator) Description(_ context.Context) string {
	return "At least one of certificate, key, certificate_authorities, or client_auth must be set"
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
		if value, ok := attrs[name]; ok && !value.IsNull() {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		sslAtLeastOneSubattrSummary,
		sslAtLeastOneSubattrMessage,
	)
}
