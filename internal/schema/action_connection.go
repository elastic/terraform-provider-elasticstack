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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetEsActionConnectionBlock returns the elasticsearch_connection block for
// provider-defined actions, mirroring GetEsEphemeralConnectionBlock for ephemeral resources.
func GetEsActionConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)
	caFilePath := path.MatchRelative().AtParent().AtName(attrCAFile)
	caDataPath := path.MatchRelative().AtParent().AtName(attrCAData)
	certFilePath := path.MatchRelative().AtParent().AtName(attrCertFile)
	certDataPath := path.MatchRelative().AtParent().AtName(attrCertData)
	keyFilePath := path.MatchRelative().AtParent().AtName(attrKeyFile)
	keyDataPath := path.MatchRelative().AtParent().AtName(attrKeyData)

	return schema.ListNestedBlock{
		MarkdownDescription: descESConnectionBlock,
		Description:         descESConnectionBlock,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrUsername: schema.StringAttribute{
					MarkdownDescription: descUsername,
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				attrPassword: schema.StringAttribute{
					MarkdownDescription: descPassword,
					Optional:            true,
					WriteOnly:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrAPIKey: schema.StringAttribute{
					MarkdownDescription: descAPIKey,
					Optional:            true,
					WriteOnly:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: schema.StringAttribute{
					MarkdownDescription: descBearerToken,
					Optional:            true,
					WriteOnly:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				attrESClientAuthentication: schema.StringAttribute{
					MarkdownDescription: descESClientAuthentication,
					Optional:            true,
					WriteOnly:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
						stringvalidator.AlsoRequires(bearerTokenPath),
					},
				},
				attrEndpoints: schema.ListAttribute{
					MarkdownDescription: descEndpoints,
					Optional:            true,
					ElementType:         types.StringType,
				},
				attrHeaders: schema.MapAttribute{
					MarkdownDescription: descHeaders,
					Optional:            true,
					WriteOnly:           true,
					ElementType:         types.StringType,
				},
				attrInsecure: schema.BoolAttribute{
					MarkdownDescription: descInsecureTLS,
					Optional:            true,
				},
				attrCAFile: schema.StringAttribute{
					MarkdownDescription: descCAFile,
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caDataPath),
					},
				},
				attrCAData: schema.StringAttribute{
					MarkdownDescription: descCAData,
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caFilePath),
					},
				},
				attrCertFile: schema.StringAttribute{
					MarkdownDescription: descCertFile,
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyFilePath),
						stringvalidator.ConflictsWith(caDataPath, keyDataPath),
					},
				},
				attrKeyFile: schema.StringAttribute{
					MarkdownDescription: descKeyFile,
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certFilePath),
						stringvalidator.ConflictsWith(certDataPath, keyDataPath),
					},
				},
				attrCertData: schema.StringAttribute{
					MarkdownDescription: descCertData,
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
				attrKeyData: schema.StringAttribute{
					MarkdownDescription: descKeyData,
					Optional:            true,
					WriteOnly:           true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// GetKbActionConnectionBlock returns the kibana_connection block for
// provider-defined actions, mirroring GetKbEphemeralConnectionBlock for
// ephemeral resources.
func GetKbActionConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return schema.ListNestedBlock{
		MarkdownDescription: descKbConnectionBlock,
		Description:         descKbConnectionBlock,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrAPIKey: schema.StringAttribute{
					MarkdownDescription: descKbAPIKey,
					Optional:            true,
					WriteOnly:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: schema.StringAttribute{
					MarkdownDescription: descKbBearerToken,
					Optional:            true,
					WriteOnly:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				attrUsername: schema.StringAttribute{
					MarkdownDescription: descKbUsername,
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				attrPassword: schema.StringAttribute{
					MarkdownDescription: descKbPassword,
					Optional:            true,
					WriteOnly:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrEndpoints: schema.ListAttribute{
					MarkdownDescription: descKbEndpoints,
					Optional:            true,
					ElementType:         types.StringType,
				},
				attrCACerts: schema.ListAttribute{
					MarkdownDescription: descKbCACerts,
					Optional:            true,
					ElementType:         types.StringType,
				},
				attrInsecure: schema.BoolAttribute{
					MarkdownDescription: descInsecureTLS,
					Optional:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
