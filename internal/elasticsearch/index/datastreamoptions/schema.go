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

package datastreamoptions

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Block returns the Plugin Framework SingleNestedBlock for the data_stream_options block.
// Suitable for embedding inside a template { } SingleNestedBlock's Blocks map.
func Block() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: BlockDescription,
		Blocks: map[string]schema.Block{
			attrFailureStore: failureStoreBlock(),
		},
		Validators: []validator.Object{
			objectvalidator.AlsoRequires(path.MatchRelative().AtName(attrFailureStore)),
		},
	}
}

func failureStoreBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: FailureStoreBlockDescription,
		Attributes: map[string]schema.Attribute{
			attrEnabled: schema.BoolAttribute{
				MarkdownDescription: FailureStoreEnabledDescription,
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			attrLifecycle: failureStoreLifecycleBlock(),
		},
	}
}

func failureStoreLifecycleBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: FailureStoreLifecycleBlockDescription,
		Attributes: map[string]schema.Attribute{
			attrDataRetention: schema.StringAttribute{
				MarkdownDescription: FailureStoreDataRetentionDescription,
				Optional:            true,
			},
		},
	}
}
