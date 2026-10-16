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

const authCredentialsSummary = "Invalid Attribute Combination"

// authCredentialsValidator validates the credential attributes of the auth
// object as one unit, covering both plain and write-only forms:
//   - username and a password form (password or password_wo) must both be
//     configured or both be absent;
//   - password and password_wo are mutually exclusive;
//   - api_key and api_key_wo are mutually exclusive;
//   - the password family conflicts with the api_key family.
//
// headers carries no dependency on any credential. Unknown values defer the
// rules that involve them, since they may resolve to null at apply.
type authCredentialsValidator struct{}

func (authCredentialsValidator) Description(_ context.Context) string {
	return "username requires password or password_wo (and vice versa); password/password_wo, api_key/api_key_wo, and the password and api_key families are mutually exclusive"
}

func (v authCredentialsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (authCredentialsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	isSet := func(name string) bool {
		v, ok := attrs[name]
		return ok && !v.IsNull() && !v.IsUnknown()
	}
	isUnknown := func(name string) bool {
		v, ok := attrs[name]
		return ok && v.IsUnknown()
	}
	anyUnknown := func(names ...string) bool {
		for _, name := range names {
			if isUnknown(name) {
				return true
			}
		}
		return false
	}
	addError := func(detail string) {
		resp.Diagnostics.AddAttributeError(req.Path, authCredentialsSummary, detail)
	}

	if isSet("password") && isSet("password_wo") {
		addError("password and password_wo are mutually exclusive.")
	}
	if isSet("api_key") && isSet("api_key_wo") {
		addError("api_key and api_key_wo are mutually exclusive.")
	}
	if (isSet("password") || isSet("password_wo")) && (isSet("api_key") || isSet("api_key_wo")) {
		addError("password or password_wo cannot be combined with api_key or api_key_wo.")
	}

	if anyUnknown("username", "password", "password_wo") {
		return
	}
	hasPassword := isSet("password") || isSet("password_wo")
	switch {
	case isSet("username") && !hasPassword:
		addError("username requires password or password_wo to be set.")
	case !isSet("username") && hasPassword:
		addError("password or password_wo requires username to be set.")
	}
}
