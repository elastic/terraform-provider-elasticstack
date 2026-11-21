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

package sslcore

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// AtLeastOneSubattrValidator returns an object validator that rejects a
// known, non-null object whose subattrNames are all null, so a practitioner
// cannot send an ssl-shaped object the Fleet API would persist as a
// present-but-empty object. Unknown values (the whole object or any child)
// intentionally defer validation: they may resolve to a set value later in
// the apply. summary and message are used verbatim for the diagnostic;
// message, with any trailing period trimmed, also becomes the validator's
// Description/MarkdownDescription text.
func AtLeastOneSubattrValidator(subattrNames []string, summary, message string) validator.Object {
	return atLeastOneSubattrValidator{
		subattrNames: subattrNames,
		summary:      summary,
		message:      message,
	}
}

type atLeastOneSubattrValidator struct {
	subattrNames []string
	summary      string
	message      string
}

func (v atLeastOneSubattrValidator) Description(_ context.Context) string {
	return strings.TrimSuffix(v.message, ".")
}

func (v atLeastOneSubattrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v atLeastOneSubattrValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	for _, name := range v.subattrNames {
		if value, ok := attrs[name]; ok && !value.IsNull() {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(req.Path, v.summary, v.message)
}
