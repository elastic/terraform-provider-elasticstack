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
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// resourceWriteAdapter captures the only meaningful differences between the
// Elasticsearch and Kibana resource write paths: client resolution,
// identity/space validation, invoking the concrete create/update callback,
// read-after-write (including its not-found error), and postRead. Everything
// else -- decoding plan/prior/config, computing the operation timeout,
// enforcing version requirements, and persisting state/timeouts -- is
// orchestrated once by [runGenericResourceWrite].
type resourceWriteAdapter[T WithResourceTimeouts, C MinVersionClient] struct {
	component Component

	// getClient resolves the scoped client for the plan model.
	getClient func(ctx context.Context, plan T) (C, diag.Diagnostics)

	// validateIdentity checks and/or derives the write identity before the
	// client is resolved. writeKey flows into invokeWrite and is available to
	// readAfterWrite for use in error messages. spaceID is a second,
	// backend-specific identity component (Kibana's space); Elasticsearch
	// adapters return "" and ignore it.
	validateIdentity func(plan T, prior *T, isUpdate bool) (writeKey string, spaceID string, diags diag.Diagnostics)

	hasWriteCallback func(isUpdate bool) bool
	hasReadFunc      func() bool

	// invokeWrite calls the concrete create/update callback and normalizes its
	// result. skipReadAfterWrite mirrors the Elasticsearch struct-level option
	// and Kibana's per-call result field under one signature.
	invokeWrite func(
		ctx context.Context, client C, plan, config T, prior *T, writeKey, spaceID string,
		private PrivateStateStorage, isUpdate bool,
	) (model T, skipReadAfterWrite bool, diags diag.Diagnostics)

	// readAfterWrite resolves the post-write read identity from the write
	// outcome's model, invokes the read callback, and reports a not-found
	// error. Only called when the write outcome does not skip read-after-write.
	readAfterWrite func(ctx context.Context, client C, model T, writeKey string) (state T, diags diag.Diagnostics)

	// postRead runs after a successful read. postReadRunsOnSkip controls
	// whether it still runs when read-after-write was skipped: Elasticsearch
	// always runs it, Kibana skips it along with the read.
	postRead           func(ctx context.Context, client C, prior, state T, private PrivateStateStorage) (T, diag.Diagnostics)
	postReadRunsOnSkip bool
}

// runGenericResourceWrite implements the shared Create/Update orchestration
// used by [ElasticsearchResource.runWrite] and [KibanaResource.runKibanaWrite]:
// guard on nil write callback, decode plan, compute the operation timeout,
// decode prior state (Update only), validate write identity, resolve the
// scoped client, enforce version requirements, decode config, invoke the
// write callback, read-after-write (unless skipped), optional postRead, then
// persist state and the timeouts attribute.
func runGenericResourceWrite[T WithResourceTimeouts, C MinVersionClient](
	ctx context.Context,
	timeoutsCfg ResourceTimeouts,
	inv resourceWriteInvocation,
	a resourceWriteAdapter[T, C],
) diag.Diagnostics {
	var diags diag.Diagnostics
	if !a.hasWriteCallback(inv.isUpdate) {
		op := envelopeWriteOpCreate
		if inv.isUpdate {
			op = envelopeWriteOpUpdate
		}
		return requireCallbackDiag(a.component, op)
	}

	var planModel T
	diags.Append(inv.plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return diags
	}

	var opTimeout time.Duration
	var timeoutDiags diag.Diagnostics
	if inv.isUpdate {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Update(ctx, timeoutsCfg.UpdateOrDefault())
	} else {
		opTimeout, timeoutDiags = planModel.GetTimeouts().Create(ctx, timeoutsCfg.CreateOrDefault())
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

	writeKey, spaceID, idDiags := a.validateIdentity(planModel, priorPtr, inv.isUpdate)
	diags.Append(idDiags...)
	if diags.HasError() {
		return diags
	}

	client, connDiags := a.getClient(ctx, planModel)
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &planModel); vDiags.HasError() {
		diags.Append(vDiags...)
		return diags
	}

	if !a.hasReadFunc() {
		return requireReadFuncDiag(a.component)
	}

	var configModel T
	diags.Append(inv.config.Get(ctx, &configModel)...)
	if diags.HasError() {
		return diags
	}

	writtenModel, skipReadAfterWrite, callDiags := a.invokeWrite(ctx, client, planModel, configModel, priorPtr, writeKey, spaceID, inv.privateState, inv.isUpdate)
	diags.Append(callDiags...)
	if diags.HasError() {
		return diags
	}

	var stateModel T
	if skipReadAfterWrite {
		stateModel = writtenModel
	} else {
		var readDiags diag.Diagnostics
		stateModel, readDiags = a.readAfterWrite(ctx, client, writtenModel, writeKey)
		diags.Append(readDiags...)
		if diags.HasError() {
			return diags
		}
	}

	if a.postRead != nil && (!skipReadAfterWrite || a.postReadRunsOnSkip) {
		var prDiags diag.Diagnostics
		stateModel, prDiags = a.postRead(ctx, client, planModel, stateModel, inv.privateState)
		diags.Append(prDiags...)
		if diags.HasError() {
			return diags
		}
	}

	preserveModelTimeouts(&stateModel, planModel.GetTimeouts())
	diags.Append(inv.outState.Set(ctx, &stateModel)...)
	if diags.HasError() {
		return diags
	}

	diags.Append(inv.outState.SetAttribute(ctx, path.Root(attrTimeouts), planModel.GetTimeouts())...)
	return diags
}
