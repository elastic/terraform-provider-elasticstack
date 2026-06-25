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

// esConnectionBlockSpec returns the canonical connectionBlockSpec for the
// elasticsearch_connection block. It is the single source of truth from which
// the managed-resource, ephemeral-resource, action-resource blocks (and the
// object attribute-type map) are generated. Adding a new connection attribute
// here automatically propagates it to every entity kind.
func esConnectionBlockSpec() connectionBlockSpec {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)
	caFilePath := path.MatchRelative().AtParent().AtName(attrCAFile)
	caDataPath := path.MatchRelative().AtParent().AtName(attrCAData)
	caFingerprintPath := path.MatchRelative().AtParent().AtName(attrCAFingerprint)
	certFilePath := path.MatchRelative().AtParent().AtName(attrCertFile)
	certDataPath := path.MatchRelative().AtParent().AtName(attrCertData)
	keyFilePath := path.MatchRelative().AtParent().AtName(attrKeyFile)
	keyDataPath := path.MatchRelative().AtParent().AtName(attrKeyData)

	return connectionBlockSpec{
		description: descESConnectionBlock,
		attrs: []connAttrSpec{
			{
				name:        attrUsername,
				description: descUsername,
				kind:        connAttrString,
				validators:  []validator.String{stringvalidator.AlsoRequires(passwordPath)},
			},
			{
				name:        attrPassword,
				description: descPassword,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators:  []validator.String{stringvalidator.AlsoRequires(usernamePath)},
			},
			{
				name:        attrAPIKey,
				description: descAPIKey,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
				},
			},
			{
				name:        attrBearerToken,
				description: descBearerToken,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
				},
			},
			{
				name:        attrESClientAuthentication,
				description: descESClientAuthentication,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					stringvalidator.AlsoRequires(bearerTokenPath),
				},
			},
			{
				name:        attrEndpoints,
				description: descEndpoints,
				kind:        connAttrList,
				sensitive:   true,
				// writeOnly is deliberately false: action resources read endpoints back.
			},
			{
				name:        attrHeaders,
				description: descHeaders,
				kind:        connAttrMap,
				sensitive:   true,
				writeOnly:   true,
			},
			{
				name:        attrInsecure,
				description: descInsecureTLS,
				kind:        connAttrBool,
			},
			{
				name:        attrCAFile,
				description: descCAFile,
				kind:        connAttrString,
				validators: []validator.String{
					stringvalidator.ConflictsWith(caDataPath, caFingerprintPath),
				},
			},
			{
				name:        attrCAData,
				description: descCAData,
				kind:        connAttrString,
				validators: []validator.String{
					stringvalidator.ConflictsWith(caFilePath, caFingerprintPath),
				},
			},
			{
				name:        attrCAFingerprint,
				description: descCAFingerprint,
				kind:        connAttrString,
				validators: []validator.String{
					stringvalidator.ConflictsWith(caFilePath, caDataPath),
				},
			},
			{
				name:        attrCertFile,
				description: descCertFile,
				kind:        connAttrString,
				validators: []validator.String{
					stringvalidator.AlsoRequires(keyFilePath),
					stringvalidator.ConflictsWith(certDataPath, keyDataPath),
				},
			},
			{
				name:        attrKeyFile,
				description: descKeyFile,
				kind:        connAttrString,
				validators: []validator.String{
					stringvalidator.AlsoRequires(certFilePath),
					stringvalidator.ConflictsWith(certDataPath, keyDataPath),
				},
			},
			{
				name:        attrCertData,
				description: descCertData,
				kind:        connAttrString,
				validators: []validator.String{
					stringvalidator.AlsoRequires(keyDataPath),
					stringvalidator.ConflictsWith(certFilePath, keyFilePath),
				},
			},
			{
				name:        attrKeyData,
				description: descKeyData,
				kind:        connAttrString,
				sensitive:   true,
				writeOnly:   true,
				validators: []validator.String{
					stringvalidator.AlsoRequires(certDataPath),
					stringvalidator.ConflictsWith(certFilePath, keyFilePath),
				},
			},
		},
	}
}
