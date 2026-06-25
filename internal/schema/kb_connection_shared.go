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

// kbConnectionBlockSpec returns the canonical connectionBlockSpec for the
// kibana_connection block. It is the single source of truth from which the
// managed-resource, ephemeral-resource, and action-resource blocks (and the
// object attribute-type map) are generated, replacing the previously
// triplicated block definitions in connection.go, ephemeral_connection.go,
// and action_connection.go.
func kbConnectionBlockSpec() connectionBlockSpec {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return connectionBlockSpec{
		description: descKbConnectionBlock,
		attrs: []connAttrSpec{
			{
				name:        attrAPIKey,
				description: descKbAPIKey,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
				},
			},
			{
				name:        attrBearerToken,
				description: descKbBearerToken,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
				},
			},
			{
				name:        attrUsername,
				description: descKbUsername,
				kind:        connAttrString,
				validators:  []validator.String{stringvalidator.AlsoRequires(passwordPath)},
			},
			{
				name:        attrPassword,
				description: descKbPassword,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators:  []validator.String{stringvalidator.AlsoRequires(usernamePath)},
			},
			{
				name:        attrEndpoints,
				description: descKbEndpoints,
				kind:        connAttrList,
				sensitive:   true,
				// writeOnly is deliberately false: action resources read endpoints back.
			},
			{
				name:        attrCACerts,
				description: descKbCACerts,
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
