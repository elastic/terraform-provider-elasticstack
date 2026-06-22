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
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type esConnAttrKind int

const (
	esConnAttrKindString esConnAttrKind = iota
	esConnAttrKindBool
	esConnAttrKindList
	esConnAttrKindMap
)

// esConnAttrDef describes a single attribute in the elasticsearch_connection block.
// Three adapter functions (buildFWESConnectionAttributes,
// buildEphemeralESConnectionAttributes, buildActionESConnectionAttributes) convert
// these descriptors to the appropriate framework-specific schema attribute types.
type esConnAttrDef struct {
	description string
	kind        esConnAttrKind
	// sensitive is set to Sensitive on managed/ephemeral resources.
	sensitive bool
	// writeOnly is set to WriteOnly on action resources.
	writeOnly  bool
	validators []validator.String
}

// esConnectionAttrDefs returns the canonical attribute definitions for the
// elasticsearch_connection block. Adding a new connection attribute here
// automatically propagates it to managed, ephemeral, and action resource types.
func esConnectionAttrDefs() map[string]esConnAttrDef {
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

	return map[string]esConnAttrDef{
		attrUsername: {
			description: descUsername,
			kind:        esConnAttrKindString,
			validators:  []validator.String{stringvalidator.AlsoRequires(passwordPath)},
		},
		attrPassword: {
			description: descPassword,
			kind:        esConnAttrKindString,
			sensitive:   true,
			writeOnly:   true,
			validators:  []validator.String{stringvalidator.AlsoRequires(usernamePath)},
		},
		attrAPIKey: {
			description: descAPIKey,
			kind:        esConnAttrKindString,
			sensitive:   true,
			writeOnly:   true,
			validators: []validator.String{
				stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
			},
		},
		attrBearerToken: {
			description: descBearerToken,
			kind:        esConnAttrKindString,
			sensitive:   true,
			writeOnly:   true,
			validators: []validator.String{
				stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
			},
		},
		attrESClientAuthentication: {
			description: descESClientAuthentication,
			kind:        esConnAttrKindString,
			sensitive:   true,
			writeOnly:   true,
			validators: []validator.String{
				stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
				stringvalidator.AlsoRequires(bearerTokenPath),
			},
		},
		attrEndpoints: {
			description: descEndpoints,
			kind:        esConnAttrKindList,
			sensitive:   true,
			// writeOnly is deliberately false: action resources read endpoints back.
		},
		attrHeaders: {
			description: descHeaders,
			kind:        esConnAttrKindMap,
			sensitive:   true,
			writeOnly:   true,
		},
		attrInsecure: {
			description: descInsecureTLS,
			kind:        esConnAttrKindBool,
		},
		attrCAFile: {
			description: descCAFile,
			kind:        esConnAttrKindString,
			validators: []validator.String{
				stringvalidator.ConflictsWith(caDataPath, caFingerprintPath),
			},
		},
		attrCAData: {
			description: descCAData,
			kind:        esConnAttrKindString,
			validators: []validator.String{
				stringvalidator.ConflictsWith(caFilePath, caFingerprintPath),
			},
		},
		attrCAFingerprint: {
			description: descCAFingerprint,
			kind:        esConnAttrKindString,
			validators: []validator.String{
				stringvalidator.ConflictsWith(caFilePath, caDataPath),
			},
		},
		attrCertFile: {
			description: descCertFile,
			kind:        esConnAttrKindString,
			validators: []validator.String{
				stringvalidator.AlsoRequires(keyFilePath),
				stringvalidator.ConflictsWith(caDataPath, keyDataPath),
			},
		},
		attrKeyFile: {
			description: descKeyFile,
			kind:        esConnAttrKindString,
			validators: []validator.String{
				stringvalidator.AlsoRequires(certFilePath),
				stringvalidator.ConflictsWith(certDataPath, keyDataPath),
			},
		},
		attrCertData: {
			description: descCertData,
			kind:        esConnAttrKindString,
			validators: []validator.String{
				stringvalidator.AlsoRequires(keyDataPath),
				stringvalidator.ConflictsWith(certFilePath, keyFilePath),
			},
		},
		attrKeyData: {
			description: descKeyData,
			kind:        esConnAttrKindString,
			sensitive:   true,
			writeOnly:   true,
			validators: []validator.String{
				stringvalidator.AlsoRequires(certDataPath),
				stringvalidator.ConflictsWith(certFilePath, keyFilePath),
			},
		},
	}
}

// buildFWESConnectionAttributes converts the shared esConnAttrDef descriptors into
// managed-resource (provider/schema) attribute types. Sensitive fields use Sensitive: true.
func buildFWESConnectionAttributes() map[string]fwschema.Attribute {
	defs := esConnectionAttrDefs()
	attrs := make(map[string]fwschema.Attribute, len(defs))
	for name, def := range defs {
		switch def.kind {
		case esConnAttrKindString:
			attrs[name] = fwschema.StringAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				Sensitive:           def.sensitive,
				Validators:          def.validators,
			}
		case esConnAttrKindBool:
			attrs[name] = fwschema.BoolAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
			}
		case esConnAttrKindList:
			attrs[name] = fwschema.ListAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				Sensitive:           def.sensitive,
				ElementType:         types.StringType,
			}
		case esConnAttrKindMap:
			attrs[name] = fwschema.MapAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				Sensitive:           def.sensitive,
				ElementType:         types.StringType,
			}
		}
	}
	return attrs
}

// buildEphemeralESConnectionAttributes converts the shared esConnAttrDef descriptors
// into ephemeral-resource (ephemeral/schema) attribute types. Sensitive fields use
// Sensitive: true.
func buildEphemeralESConnectionAttributes() map[string]ephemeralschema.Attribute {
	defs := esConnectionAttrDefs()
	attrs := make(map[string]ephemeralschema.Attribute, len(defs))
	for name, def := range defs {
		switch def.kind {
		case esConnAttrKindString:
			attrs[name] = ephemeralschema.StringAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				Sensitive:           def.sensitive,
				Validators:          def.validators,
			}
		case esConnAttrKindBool:
			attrs[name] = ephemeralschema.BoolAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
			}
		case esConnAttrKindList:
			attrs[name] = ephemeralschema.ListAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				Sensitive:           def.sensitive,
				ElementType:         types.StringType,
			}
		case esConnAttrKindMap:
			attrs[name] = ephemeralschema.MapAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				Sensitive:           def.sensitive,
				ElementType:         types.StringType,
			}
		}
	}
	return attrs
}

// buildActionESConnectionAttributes converts the shared esConnAttrDef descriptors
// into action-resource (action/schema) attribute types. Sensitive fields use
// WriteOnly: true instead of Sensitive.
func buildActionESConnectionAttributes() map[string]actionschema.Attribute {
	defs := esConnectionAttrDefs()
	attrs := make(map[string]actionschema.Attribute, len(defs))
	for name, def := range defs {
		switch def.kind {
		case esConnAttrKindString:
			attrs[name] = actionschema.StringAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				WriteOnly:           def.writeOnly,
				Validators:          def.validators,
			}
		case esConnAttrKindBool:
			attrs[name] = actionschema.BoolAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
			}
		case esConnAttrKindList:
			attrs[name] = actionschema.ListAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				ElementType:         types.StringType,
			}
		case esConnAttrKindMap:
			attrs[name] = actionschema.MapAttribute{
				MarkdownDescription: def.description,
				Optional:            true,
				WriteOnly:           def.writeOnly,
				ElementType:         types.StringType,
			}
		}
	}
	return attrs
}
