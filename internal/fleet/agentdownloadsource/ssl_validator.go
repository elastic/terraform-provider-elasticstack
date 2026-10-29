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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const sslAtLeastOneSubattrSummary = "Empty ssl configuration"

const sslAtLeastOneSubattrMessage = "At least one of certificate_authorities, certificate, key, or key_wo must be set."

// keyAttributeName is the shared name of the ssl private-key attribute and
// the headers header-name attribute; golangci requires a constant at three
// occurrences.
const keyAttributeName = "key"

var sslSubattrNames = []string{"certificate_authorities", "certificate", keyAttributeName, "key_wo"}

// sslAtLeastOneSubattrValidator returns an object validator rejecting an ssl object whose
// sub-attributes are all null, so a practitioner cannot send an ssl object the Fleet API would
// persist as a present-but-empty object. Unknown values (the whole object or any child)
// intentionally defer validation: they may resolve to a set value later in the apply.
func sslAtLeastOneSubattrValidator() validator.Object {
	return validators.AtLeastOneOfNestedAttrsValidator(validators.AtLeastOneOfNestedAttrsOpts{
		AttrNames:   sslSubattrNames,
		Summary:     sslAtLeastOneSubattrSummary,
		Detail:      sslAtLeastOneSubattrMessage,
		Description: "At least one of certificate_authorities, certificate, key, or key_wo must be set",
	})
}
