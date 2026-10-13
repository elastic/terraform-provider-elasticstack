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

package securityexceptionitem

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/stretchr/testify/require"
)

func TestSchema_CommentsComputedOnly(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	require.EqualValues(t, 1, s.Version, "the comments schema expansion must be schema version 1")

	comments, ok := s.Attributes["comments"].(schema.ListNestedAttribute)
	require.True(t, ok, "comments must be a list nested attribute")
	require.False(t, comments.Optional, "comments must not be settable in configuration")
	require.True(t, comments.Computed, "comments must be computed")

	nestedAttrs := comments.NestedObject.Attributes
	require.Equal(t, map[string]bool{
		"id":         true,
		"comment":    true,
		"created_at": true,
		"created_by": true,
	}, computedFlags(nestedAttrs))
}

func TestSchema_RequiredComments(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())
	requiredComments, ok := s.Attributes["required_comments"].(schema.ListAttribute)
	require.True(t, ok, "required_comments must be a list attribute")
	require.True(t, requiredComments.Optional)
	require.NotEmpty(t, requiredComments.Validators, "required_comments must carry plan-time validators")
}

func TestSchema_CommentPlanModifierWiring(t *testing.T) {
	t.Parallel()

	s := getSchema(context.Background())

	comments := s.Attributes["comments"].(schema.ListNestedAttribute)
	require.Equal(t, []planmodifier.List{unknownWhenRequiredCommentMissing{}}, comments.PlanModifiers)

	requiredComments := s.Attributes["required_comments"].(schema.ListAttribute)
	require.Empty(t, requiredComments.PlanModifiers, "required_comments must not carry plan modifiers")
}

func computedFlags(attrs map[string]schema.Attribute) map[string]bool {
	flags := make(map[string]bool, len(attrs))
	for name, attr := range attrs {
		switch a := attr.(type) {
		case schema.StringAttribute:
			flags[name] = a.Computed
		default:
			flags[name] = false
		}
	}
	return flags
}
