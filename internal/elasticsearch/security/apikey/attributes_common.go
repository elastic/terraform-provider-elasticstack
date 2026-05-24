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

package apikey

import (
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Attribute descriptions shared between the resource and ephemeral resource
// schemas. Per-flavor overrides (e.g. ephemeral's expiration that mentions
// invalidate_on_close) live next to the schema that needs them.
const (
	NameDescription                = "Specifies the name for this API key."
	TypeDescription                = "The type of API key. Valid values are 'rest' (default) and 'cross_cluster'. Cross-cluster API keys are used for cross-cluster search and replication."
	RoleDescriptorsDescription     = "Role descriptors for this API key."
	ExpirationDescription          = "Expiration time for the API key. By default, API keys never expire."
	ExpirationTimestampDescription = "Expiration time in milliseconds for the API key. By default, API keys never expire."
	MetadataDescription            = "Arbitrary metadata that you want to associate with the API key."
	AccessDescription              = "Access configuration for cross-cluster API keys. Only applicable when type is 'cross_cluster'."
	KeyIDDescription               = "Unique id for this API key."
	APIKeyDescription              = "Generated API Key."
	EncodedDescription             = "API key credentials which is the Base64-encoding of the UTF-8 representation of the id and api_key joined by a colon (:)."

	AccessSearchDescription                 = "A list of search configurations for which the cross-cluster API key will have search privileges."
	AccessReplicationDescription            = "A list of replication configurations for which the cross-cluster API key will have replication privileges."
	AccessSearchNamesDescription            = "A list of index patterns for search."
	AccessReplicationNamesDescription       = "A list of index patterns for replication."
	AccessFieldSecurityDescription          = "Field-level security configuration in JSON format."
	AccessQueryDescription                  = "Query to filter documents for search operations in JSON format."
	AccessAllowRestrictedIndicesDescription = "Whether to allow access to restricted indices."
)

// NameValidators returns the validator set for the `name` attribute. Shared
// between the resource and ephemeral resource schemas.
func NameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 1024),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[[:graph:]]([[:graph:]]| )*[[:graph:]]$|^[[:graph:]]$`),
			APIKeyNameInvalidMessage,
		),
	}
}

// TypeValidators returns the validator set for the `type` attribute.
func TypeValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf(DefaultAPIKeyType, CrossClusterAPIKeyType),
	}
}

// RoleDescriptorsCustomType returns the JSON-with-defaults custom type used for
// the `role_descriptors` attribute, configured with the defaults populator.
func RoleDescriptorsCustomType() basetypes.StringTypable {
	return customtypes.NewJSONWithDefaultsType(PopulateRoleDescriptorsDefaults)
}

// RoleDescriptorsValidators returns the validator set for the
// `role_descriptors` attribute.
func RoleDescriptorsValidators() []validator.String {
	return []validator.String{
		RequiresType(DefaultAPIKeyType),
	}
}

// AccessValidators returns the validator set for the `access` attribute.
func AccessValidators() []validator.Object {
	return []validator.Object{
		RequiresType(CrossClusterAPIKeyType),
	}
}

// AccessAttributesResource returns the nested `access` attribute map typed for
// a resource schema.
func AccessAttributesResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"search": schema.ListNestedAttribute{
			Description: AccessSearchDescription,
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"names": schema.ListAttribute{
						Description: AccessSearchNamesDescription,
						Required:    true,
						ElementType: types.StringType,
					},
					"field_security": schema.StringAttribute{
						Description: AccessFieldSecurityDescription,
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"query": schema.StringAttribute{
						Description: AccessQueryDescription,
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"allow_restricted_indices": schema.BoolAttribute{
						Description: AccessAllowRestrictedIndicesDescription,
						Optional:    true,
					},
				},
			},
		},
		"replication": schema.ListNestedAttribute{
			Description: AccessReplicationDescription,
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"names": schema.ListAttribute{
						Description: AccessReplicationNamesDescription,
						Required:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

// AccessAttributesEphemeral returns the nested `access` attribute map typed
// for an ephemeral resource schema.
func AccessAttributesEphemeral() map[string]eschema.Attribute {
	return map[string]eschema.Attribute{
		"search": eschema.ListNestedAttribute{
			Description: AccessSearchDescription,
			Optional:    true,
			NestedObject: eschema.NestedAttributeObject{
				Attributes: map[string]eschema.Attribute{
					"names": eschema.ListAttribute{
						Description: AccessSearchNamesDescription,
						Required:    true,
						ElementType: types.StringType,
					},
					"field_security": eschema.StringAttribute{
						Description: AccessFieldSecurityDescription,
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"query": eschema.StringAttribute{
						Description: AccessQueryDescription,
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"allow_restricted_indices": eschema.BoolAttribute{
						Description: AccessAllowRestrictedIndicesDescription,
						Optional:    true,
					},
				},
			},
		},
		"replication": eschema.ListNestedAttribute{
			Description: AccessReplicationDescription,
			Optional:    true,
			NestedObject: eschema.NestedAttributeObject{
				Attributes: map[string]eschema.Attribute{
					"names": eschema.ListAttribute{
						Description: AccessReplicationNamesDescription,
						Required:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}
