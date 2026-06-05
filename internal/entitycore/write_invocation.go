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

package entitycore

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// resourceWriteInvocation carries the framework state objects needed by both
// the Elasticsearch and Kibana write paths. All fields are concrete framework
// types, so no type parameter is required. Both runWrite (resource_envelope.go)
// and runKibanaWrite (kibana_resource_envelope.go) consume this struct.
type resourceWriteInvocation struct {
	plan         tfsdk.Plan
	priorState   *tfsdk.State
	config       tfsdk.Config
	outState     *tfsdk.State
	privateState PrivateStateStorage
	isUpdate     bool
}

// requireReadFuncDiag returns an error diagnostic when the read callback for an
// envelope is nil. component ("elasticsearch", "kibana", …) is capitalized to
// form the human-readable envelope name used in both the summary and detail.
func requireReadFuncDiag(component Component) diag.Diagnostics {
	return requireCallbackDiag(component, "read")
}

// requireDeleteFuncDiag returns an error diagnostic when the delete callback
// for an envelope is nil.
func requireDeleteFuncDiag(component Component) diag.Diagnostics {
	return requireCallbackDiag(component, "delete")
}

func requireCallbackDiag(component Component, callback string) diag.Diagnostics {
	var diags diag.Diagnostics
	name := string(component)
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	diags.AddError(
		name+" envelope configuration error",
		"The "+callback+" callback passed via "+name+"ResourceOptions must not be nil.",
	)
	return diags
}

// runWriteCommon executes the shared Create/Update orchestration flow that is
// identical between [ElasticsearchResource.runWrite] and
// [KibanaResource.runKibanaWrite]:
//
//  1. Verify the appropriate write callback (create or update) is non-nil.
//  2. Decode the plan model.
//  3. Resolve and apply the operation timeout to ctx.
//  4. Decode prior state (Update only).
//  5. Decode the config model.
//  6. Invoke body, which handles client-specific work: identity resolution,
//     client creation, version enforcement, write dispatch, read-after-write,
//     and postRead.
//  7. Preserve timeouts on the returned state model and persist it to outState.
//
// body receives the deadline-constrained ctx, the decoded plan model, a pointer
// to the decoded prior model (nil on Create), and the decoded config model.
// It returns the fully-resolved state model and any diagnostics.
func runWriteCommon[T WithResourceTimeouts](
	ctx context.Context,
	component Component,
	ts ResourceTimeouts,
	inv resourceWriteInvocation,
	hasCreateFunc, hasUpdateFunc bool,
	body func(ctx context.Context, planModel T, priorPtr *T, configModel T) (T, diag.Diagnostics),
) diag.Diagnostics {
	var diags diag.Diagnostics
	if (inv.isUpdate && !hasUpdateFunc) || (!inv.isUpdate && !hasCreateFunc) {
		op := envelopeWriteOpCreate
		if inv.isUpdate {
			op = envelopeWriteOpUpdate
		}
		return requireCallbackDiag(component, op)
	}

	var planModel T
	diags.Append(inv.plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return diags
	}

	var opTimeout time.Duration
	var timeoutDiags diag.Diagnostics
	if inv.isUpdate {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Update(ctx, ts.UpdateOrDefault())
	} else {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Create(ctx, ts.CreateOrDefault())
	}
	diags.Append(timeoutDiags...)
	if diags.HasError() {
		return diags
	}
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var priorPtr *T
	if inv.isUpdate && inv.priorState != nil {
		var priorModel T
		diags.Append(inv.priorState.Get(ctx, &priorModel)...)
		if diags.HasError() {
			return diags
		}
		priorPtr = &priorModel
	}

	var configModel T
	diags.Append(inv.config.Get(ctx, &configModel)...)
	if diags.HasError() {
		return diags
	}

	stateModel, bodyDiags := body(ctx, planModel, priorPtr, configModel)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return diags
	}

	preserveModelTimeouts(&stateModel, planModel.GetTimeouts())
	diags.Append(inv.outState.Set(ctx, &stateModel)...)
	if diags.HasError() {
		return diags
	}
	diags.Append(inv.outState.SetAttribute(ctx, path.Root(attrTimeouts), planModel.GetTimeouts())...)
	return diags
}
