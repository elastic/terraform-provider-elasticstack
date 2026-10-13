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

package typeutils

import "github.com/hashicorp/terraform-plugin-framework/attr"

// ElementTypeOf asserts t to attr.TypeWithElementType and returns its ElementType().
// Used by schema.go files to pull the element type out of a list/set attribute type,
// e.g. typeutils.ElementTypeOf(schemaAttr.GetType()), for use when constructing
// types.ListValue/types.SetValue. Panics if t does not implement
// attr.TypeWithElementType.
func ElementTypeOf(t attr.Type) attr.Type {
	return t.(attr.TypeWithElementType).ElementType()
}
