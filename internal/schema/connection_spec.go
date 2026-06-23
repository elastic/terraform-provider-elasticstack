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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connAttrKind enumerates the attribute types supported by connection blocks.
type connAttrKind int

const (
	connAttrString connAttrKind = iota
	connAttrBool
	connAttrList
	connAttrMap
)

// connAttrSpec describes a single attribute in a connection block in a
// schema-package-agnostic way. A connectionBlockSpec converts these
// descriptors into the appropriate Terraform Plugin Framework attribute type
// for managed resources, ephemeral resources, and provider-defined actions.
type connAttrSpec struct {
	name string
	// description is set as MarkdownDescription (and plain Description at the
	// block level) on every variant.
	description string
	kind        connAttrKind
	// sensitive is set to Sensitive: true on managed-resource and ephemeral
	// variants. It is ignored by the action variant, which uses WriteOnly.
	sensitive bool
	// writeOnly is set to WriteOnly: true on the action variant. It is ignored
	// by the managed and ephemeral variants, which use Sensitive.
	writeOnly  bool
	validators []validator.String
}

// connectionBlockSpec is the single source of truth for a connection block
// (e.g. elasticsearch_connection, kibana_connection). It generates the block
// for every Terraform entity kind that needs one, so adding a new attribute or
// a new entity kind only requires touching this spec (or the generator).
type connectionBlockSpec struct {
	description string
	attrs       []connAttrSpec
}

// connAttrFactory builds a single connection attribute using the target
// Terraform Plugin Framework schema package. Implementations exist for managed
// resources (provider/schema), ephemeral resources (ephemeral/schema), and
// provider-defined actions (action/schema).
type connAttrFactory[T any] interface {
	stringAttr(s connAttrSpec) T
	boolAttr(s connAttrSpec) T
	listAttr(s connAttrSpec) T
	mapAttr(s connAttrSpec) T
}

// buildConnAttributes converts the spec's attribute descriptors into
// framework-specific attributes using the supplied factory. It is a free
// function because Go does not permit methods to have type parameters.
func buildConnAttributes[T any](spec connectionBlockSpec, factory connAttrFactory[T]) map[string]T {
	attrs := make(map[string]T, len(spec.attrs))
	for _, a := range spec.attrs {
		switch a.kind {
		case connAttrString:
			attrs[a.name] = factory.stringAttr(a)
		case connAttrBool:
			attrs[a.name] = factory.boolAttr(a)
		case connAttrList:
			attrs[a.name] = factory.listAttr(a)
		case connAttrMap:
			attrs[a.name] = factory.mapAttr(a)
		default:
			// connAttrKind is an internal type and the specs are the only source
			// of definitions, so an unhandled kind is a programming error.
			panic(fmt.Sprintf("unsupported connection attribute kind: %v", a.kind))
		}
	}
	return attrs
}

// attrType returns the framework attr.Type for a connection attribute, used to
// derive object attribute-type maps (e.g. for null list values and state
// upgraders) without re-declaring the attribute set by hand.
func (s connAttrSpec) attrType() attr.Type {
	switch s.kind {
	case connAttrString:
		return types.StringType
	case connAttrBool:
		return types.BoolType
	case connAttrList:
		return types.ListType{ElemType: types.StringType}
	case connAttrMap:
		return types.MapType{ElemType: types.StringType}
	default:
		panic(fmt.Sprintf("unsupported connection attribute kind: %v", s.kind))
	}
}

// attrTypes returns the attribute-type map for the block's nested object. It is
// the single derivation point for connection-block object types, replacing the
// previously hand-maintained fallback maps.
func (s connectionBlockSpec) attrTypes() map[string]attr.Type {
	out := make(map[string]attr.Type, len(s.attrs))
	for _, a := range s.attrs {
		out[a.name] = a.attrType()
	}
	return out
}

// fwBlock builds the managed-resource (provider/schema) connection block.
// Sensitive fields use Sensitive: true. The same block shape is reused by data
// sources, which inject it via the entitycore envelope.
func (s connectionBlockSpec) fwBlock() fwschema.Block {
	return fwschema.ListNestedBlock{
		MarkdownDescription: s.description,
		Description:         s.description,
		NestedObject: fwschema.NestedBlockObject{
			Attributes: buildConnAttributes(s, fwConnAttrFactory{}),
		},
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
	}
}

// ephemeralBlock builds the ephemeral-resource (ephemeral/schema) connection
// block. Sensitive fields use Sensitive: true.
func (s connectionBlockSpec) ephemeralBlock() ephemeralschema.Block {
	return ephemeralschema.ListNestedBlock{
		MarkdownDescription: s.description,
		Description:         s.description,
		NestedObject: ephemeralschema.NestedBlockObject{
			Attributes: buildConnAttributes(s, ephemeralConnAttrFactory{}),
		},
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
	}
}

// actionBlock builds the action-resource (action/schema) connection block.
// Sensitive fields use WriteOnly: true instead of Sensitive.
func (s connectionBlockSpec) actionBlock() actionschema.Block {
	return actionschema.ListNestedBlock{
		MarkdownDescription: s.description,
		Description:         s.description,
		NestedObject: actionschema.NestedBlockObject{
			Attributes: buildConnAttributes(s, actionConnAttrFactory{}),
		},
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
	}
}

// fwConnAttrFactory builds managed-resource (provider/schema) attributes.
type fwConnAttrFactory struct{}

func (fwConnAttrFactory) stringAttr(s connAttrSpec) fwschema.Attribute {
	return fwschema.StringAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		Sensitive:           s.sensitive,
		Validators:          s.validators,
	}
}

func (fwConnAttrFactory) boolAttr(s connAttrSpec) fwschema.Attribute {
	return fwschema.BoolAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
	}
}

func (fwConnAttrFactory) listAttr(s connAttrSpec) fwschema.Attribute {
	return fwschema.ListAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		Sensitive:           s.sensitive,
		ElementType:         types.StringType,
	}
}

func (fwConnAttrFactory) mapAttr(s connAttrSpec) fwschema.Attribute {
	return fwschema.MapAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		Sensitive:           s.sensitive,
		ElementType:         types.StringType,
	}
}

// ephemeralConnAttrFactory builds ephemeral-resource (ephemeral/schema) attributes.
type ephemeralConnAttrFactory struct{}

func (ephemeralConnAttrFactory) stringAttr(s connAttrSpec) ephemeralschema.Attribute {
	return ephemeralschema.StringAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		Sensitive:           s.sensitive,
		Validators:          s.validators,
	}
}

func (ephemeralConnAttrFactory) boolAttr(s connAttrSpec) ephemeralschema.Attribute {
	return ephemeralschema.BoolAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
	}
}

func (ephemeralConnAttrFactory) listAttr(s connAttrSpec) ephemeralschema.Attribute {
	return ephemeralschema.ListAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		Sensitive:           s.sensitive,
		ElementType:         types.StringType,
	}
}

func (ephemeralConnAttrFactory) mapAttr(s connAttrSpec) ephemeralschema.Attribute {
	return ephemeralschema.MapAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		Sensitive:           s.sensitive,
		ElementType:         types.StringType,
	}
}

// actionConnAttrFactory builds action-resource (action/schema) attributes.
type actionConnAttrFactory struct{}

func (actionConnAttrFactory) stringAttr(s connAttrSpec) actionschema.Attribute {
	return actionschema.StringAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		WriteOnly:           s.writeOnly,
		Validators:          s.validators,
	}
}

func (actionConnAttrFactory) boolAttr(s connAttrSpec) actionschema.Attribute {
	return actionschema.BoolAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
	}
}

func (actionConnAttrFactory) listAttr(s connAttrSpec) actionschema.Attribute {
	return actionschema.ListAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		ElementType:         types.StringType,
	}
}

func (actionConnAttrFactory) mapAttr(s connAttrSpec) actionschema.Attribute {
	return actionschema.MapAttribute{
		MarkdownDescription: s.description,
		Optional:            true,
		WriteOnly:           s.writeOnly,
		ElementType:         types.StringType,
	}
}
