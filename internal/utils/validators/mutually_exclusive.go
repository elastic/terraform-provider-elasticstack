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

package validators

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// MutuallyExclusiveStringValidator returns a single-element validator slice declaring this
// string attribute conflicts with the named sibling on the parent object.
func MutuallyExclusiveStringValidator(siblingName string) []validator.String {
	return []validator.String{
		stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(siblingName)),
	}
}

// MutuallyExclusiveListValidator is the list-attribute counterpart to MutuallyExclusiveStringValidator.
func MutuallyExclusiveListValidator(siblingName string) []validator.List {
	return []validator.List{
		listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(siblingName)),
	}
}

// MutuallyExclusiveObjectValidator is the object-attribute counterpart to MutuallyExclusiveStringValidator.
func MutuallyExclusiveObjectValidator(siblingName string) []validator.Object {
	return []validator.Object{
		objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(siblingName)),
	}
}
