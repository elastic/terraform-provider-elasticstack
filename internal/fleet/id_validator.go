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

package fleet

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const idValidatorDescription = `Must be 1-255 characters and must not contain path separators ("/"), traversal sequences (".."), or reserved keys ("__proto__", "constructor", "prototype").`

var (
	_ validator.String = idValidator{}

	idReservedSubstrings = [...]string{"__proto__", "constructor", "prototype"}
)

type idValidator struct {
	attributeName string
}

// IDValidator returns a reusable Fleet ID validator for the given Terraform attribute name.
func IDValidator(attributeName string) validator.String {
	return idValidator{attributeName: attributeName}
}

func (idValidator) Description(_ context.Context) string {
	return idValidatorDescription
}

func (v idValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v idValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	value := req.ConfigValue.ValueString()
	runeCount := utf8.RuneCountInString(value)
	if runeCount < 1 || runeCount > 255 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ID length",
			idLengthErrorDetail(v.attributeName),
		)
		return
	}

	if strings.Contains(value, "/") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ID",
			idPathErrorDetail(v.attributeName),
		)
		return
	}

	if strings.Contains(value, "..") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid ID",
			idTraversalErrorDetail(v.attributeName),
		)
		return
	}

	for _, reserved := range idReservedSubstrings {
		if strings.Contains(value, reserved) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid ID",
				idReservedErrorDetail(v.attributeName, reserved),
			)
			return
		}
	}
}

func idLengthErrorDetail(attributeName string) string {
	return fmt.Sprintf("%s must be between 1 and 255 characters (inclusive).", attributeName)
}

func idPathErrorDetail(attributeName string) string {
	return fmt.Sprintf(`%s must not contain path separators ("/").`, attributeName)
}

func idTraversalErrorDetail(attributeName string) string {
	return fmt.Sprintf(`%s must not contain traversal sequences ("..").`, attributeName)
}

func idReservedErrorDetail(attributeName, reserved string) string {
	return fmt.Sprintf(`%s must not contain reserved keys (%q).`, attributeName, reserved)
}
