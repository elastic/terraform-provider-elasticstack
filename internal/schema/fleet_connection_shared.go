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

package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// fleetConnectionBlockSpec returns the canonical connectionBlockSpec for the
// fleet connection block. It is the single source of truth from which the
// managed-resource block (and the object attribute-type map) is generated,
// mirroring esConnectionBlockSpec/kbConnectionBlockSpec.
func fleetConnectionBlockSpec() connectionBlockSpec {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return connectionBlockSpec{
		description: descFleetConnectionBlock,
		attrs: []connAttrSpec{
			{
				name:        attrUsername,
				description: descFleetUsername,
				kind:        connAttrString,
				validators:  []validator.String{stringvalidator.AlsoRequires(passwordPath)},
			},
			{
				name:        attrPassword,
				description: descFleetPassword,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators:  []validator.String{stringvalidator.AlsoRequires(usernamePath)},
			},
			{
				name:        attrAPIKey,
				description: descFleetAPIKey,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
				},
			},
			{
				name:        attrBearerToken,
				description: descFleetBearerToken,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
				},
			},
			{
				name:        attrEndpoint,
				description: descFleetEndpoint,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
			},
			{
				name:        attrCACerts,
				description: descFleetCACerts,
				kind:        connAttrList,
			},
			{
				name:        attrInsecure,
				description: descInsecureTLS,
				kind:        connAttrBool,
			},
		},
	}
}
