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

	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// writeCommonAdapter captures the parts of the write orchestration path that
// differ between Elasticsearch and Kibana envelopes. [runWriteCommon] handles
// the identical portions (nil-callback guard, plan/prior decode, version
// enforcement, config decode, read-after-write, postRead, state persistence);
// adapters supply the remaining client-specific pieces.
type writeCommonAdapter[T any, C MinVersionClient] struct {
	// validateIdentity validates and resolves the write identity from planModel
	// and (for updates) priorPtr. For Kibana it also validates spaceID.
	validateIdentity func(planModel T, priorPtr *T, isUpdate bool) (writeID string, diags diag.Diagnostics)

	// getClient resolves the scoped API client from planModel.
	getClient func(ctx context.Context, planModel T) (C, diag.Diagnostics)

	// checkReadFunc returns an error diagnostic when the read callback is nil.
	checkReadFunc func() diag.Diagnostics

	// invokeWrite constructs the concrete write request and calls the create or
	// update callback.
	invokeWrite func(ctx context.Context, client C, planModel T, priorPtr *T, configModel T, writeID string, isUpdate bool, private PrivateStateStorage) (T, diag.Diagnostics)

	// resolveReadIdentity validates the written model and returns the resource
	// ID to use for the read-after-write call.
	resolveReadIdentity func(writtenModel T, writeID string) (readID string, diags diag.Diagnostics)

	// doRead calls the read callback using the resolved identity.
	doRead func(ctx context.Context, client C, readID string, writtenModel T) (T, bool, diag.Diagnostics)

	// notFoundDetail returns the Detail string for the "Resource not found"
	// diagnostic. Receives both writeID (the original write identity) and readID
	// (the post-write resolved read identity) so each adapter can choose which
	// to surface.
	notFoundDetail func(writeID, readID string) string

	// doPostRead calls the post-read hook; the adapter wraps the nil check for
	// opts.PostRead so runWriteCommon can always call this unconditionally.
	doPostRead func(ctx context.Context, client C, priorModel T, stateModel T, private PrivateStateStorage) (T, diag.Diagnostics)
}

// runWriteCommon is the shared Create/Update orchestration kernel used by both
// [ElasticsearchResource.runWrite] and [KibanaResource.runKibanaWrite]. It
// handles the nil-callback guard, plan decode, prior-state decode, client
// resolution, version enforcement, config decode, write dispatch,
// read-after-write, postRead, and final state persistence. The pieces that
// differ between Elasticsearch and Kibana are delegated to adapter.
func runWriteCommon[T any, C MinVersionClient](
	ctx context.Context,
	inv resourceWriteInvocation,
	component Component,
	isCreateNil, isUpdateNil bool,
	adapter writeCommonAdapter[T, C],
) diag.Diagnostics {
	var diags diag.Diagnostics

	if (inv.isUpdate && isUpdateNil) || (!inv.isUpdate && isCreateNil) {
		op := "create"
		if inv.isUpdate {
			op = "update"
		}
		return requireCallbackDiag(component, op)
	}

	var planModel T
	diags.Append(inv.plan.Get(ctx, &planModel)...)
	if diags.HasError() {
		return diags
	}

	var priorPtr *T
	if inv.isUpdate && inv.priorState != nil {
		var priorModel T
		diags.Append(inv.priorState.Get(ctx, &priorModel)...)
		if diags.HasError() {
			return diags
		}
		priorPtr = &priorModel
	}

	writeID, idDiags := adapter.validateIdentity(planModel, priorPtr, inv.isUpdate)
	diags.Append(idDiags...)
	if diags.HasError() {
		return diags
	}

	client, connDiags := adapter.getClient(ctx, planModel)
	diags.Append(connDiags...)
	if diags.HasError() {
		return diags
	}

	if vDiags := EnforceVersionRequirements(ctx, client, &planModel); vDiags.HasError() {
		diags.Append(vDiags...)
		return diags
	}

	if d := adapter.checkReadFunc(); d.HasError() {
		return d
	}

	var configModel T
	diags.Append(inv.config.Get(ctx, &configModel)...)
	if diags.HasError() {
		return diags
	}

	writtenModel, callDiags := adapter.invokeWrite(ctx, client, planModel, priorPtr, configModel, writeID, inv.isUpdate, inv.privateState)
	diags.Append(callDiags...)
	if diags.HasError() {
		return diags
	}

	readID, ridDiags := adapter.resolveReadIdentity(writtenModel, writeID)
	diags.Append(ridDiags...)
	if diags.HasError() {
		return diags
	}

	stateModel, found, readDiags := adapter.doRead(ctx, client, readID, writtenModel)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	if !found {
		diags.AddError("Resource not found", adapter.notFoundDetail(writeID, readID))
		return diags
	}

	priorModel := planModel
	stateModel, prDiags := adapter.doPostRead(ctx, client, priorModel, stateModel, inv.privateState)
	diags.Append(prDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(inv.outState.Set(ctx, &stateModel)...)
	return diags
}
