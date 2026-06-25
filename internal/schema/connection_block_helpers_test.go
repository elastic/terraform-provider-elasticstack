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

	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func fwConnectionBlockAttributeNames(block fwschema.Block) map[string]struct{} {
	lb, ok := block.(fwschema.ListNestedBlock)
	if !ok {
		panic(fmt.Sprintf("connection block is %T, want ListNestedBlock", block))
	}
	return attributeNameSet(lb.NestedObject.Attributes)
}

func ephemeralConnectionBlockAttributeNames(block ephemeralschema.Block) map[string]struct{} {
	lb, ok := block.(ephemeralschema.ListNestedBlock)
	if !ok {
		panic(fmt.Sprintf("connection block is %T, want ListNestedBlock", block))
	}
	return attributeNameSet(lb.NestedObject.Attributes)
}

func actionConnectionBlockAttributeNames(block actionschema.Block) map[string]struct{} {
	lb, ok := block.(actionschema.ListNestedBlock)
	if !ok {
		panic(fmt.Sprintf("connection block is %T, want ListNestedBlock", block))
	}
	return attributeNameSet(lb.NestedObject.Attributes)
}

func attributeNameSet[T any](attrs map[string]T) map[string]struct{} {
	names := make(map[string]struct{}, len(attrs))
	for name := range attrs {
		names[name] = struct{}{}
	}
	return names
}

type tlsTrustAttributeValidatorCounts struct {
	caFile        int
	caData        int
	caFingerprint int
}

func countFWStringValidators(block fwschema.Block, attr string) int {
	lb, ok := block.(fwschema.ListNestedBlock)
	if !ok {
		panic(fmt.Sprintf("connection block is %T, want ListNestedBlock", block))
	}
	a, ok := lb.NestedObject.Attributes[attr].(fwschema.StringAttribute)
	if !ok {
		panic(fmt.Sprintf("attribute %q is %T, want StringAttribute", attr, lb.NestedObject.Attributes[attr]))
	}
	return len(a.Validators)
}

func countEphemeralStringValidators(block ephemeralschema.Block, attr string) int {
	lb, ok := block.(ephemeralschema.ListNestedBlock)
	if !ok {
		panic(fmt.Sprintf("connection block is %T, want ListNestedBlock", block))
	}
	a, ok := lb.NestedObject.Attributes[attr].(ephemeralschema.StringAttribute)
	if !ok {
		panic(fmt.Sprintf("attribute %q is %T, want StringAttribute", attr, lb.NestedObject.Attributes[attr]))
	}
	return len(a.Validators)
}

func countActionStringValidators(block actionschema.Block, attr string) int {
	lb, ok := block.(actionschema.ListNestedBlock)
	if !ok {
		panic(fmt.Sprintf("connection block is %T, want ListNestedBlock", block))
	}
	a, ok := lb.NestedObject.Attributes[attr].(actionschema.StringAttribute)
	if !ok {
		panic(fmt.Sprintf("attribute %q is %T, want StringAttribute", attr, lb.NestedObject.Attributes[attr]))
	}
	return len(a.Validators)
}
