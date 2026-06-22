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
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetEsEphemeralConnectionBlock returns the elasticsearch_connection block for
// ephemeral resources, mirroring GetEsFWConnectionBlock for managed resources.
func GetEsEphemeralConnectionBlock() schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: descESConnectionBlock,
		Description:         descESConnectionBlock,
		NestedObject: schema.NestedBlockObject{
			Attributes: buildEphemeralESConnectionAttributes(),
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// GetKbEphemeralConnectionBlock returns the kibana_connection block for
// ephemeral resources, mirroring GetKbFWConnectionBlock for managed resources.
func GetKbEphemeralConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return schema.ListNestedBlock{
		MarkdownDescription: descKbConnectionBlock,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrAPIKey: schema.StringAttribute{
					MarkdownDescription: descKbAPIKey,
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: schema.StringAttribute{
					MarkdownDescription: descKbBearerToken,
					Optional:            true,
					Sensitive:           true,
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
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrEndpoints: schema.ListAttribute{
					MarkdownDescription: descKbEndpoints,
					Optional:            true,
					Sensitive:           true,
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
