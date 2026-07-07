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

package ccr

import (
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/follow"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/putautofollowpattern"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ccr/resumefollow"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TuningParams holds the 10 common CCR tuning parameters in their Terraform type form.
type TuningParams struct {
	MaxOutstandingReadRequests    types.Int64
	MaxOutstandingWriteRequests   types.Int64
	MaxReadRequestOperationCount  types.Int64
	MaxReadRequestSize            types.String
	MaxRetryDelay                 customtypes.Duration
	MaxWriteBufferCount           types.Int64
	MaxWriteBufferSize            types.String
	MaxWriteRequestOperationCount types.Int64
	MaxWriteRequestSize           types.String
	ReadPollTimeout               customtypes.Duration
}

// ApplyToPutAutoFollowRequest sets the tuning fields on req.
// MaxOutstandingReadRequests is only sent when > 0 because the auto-follow PUT API
// rejects non-positive values (the Computed zero echoed back on read must not be
// forwarded).
func ApplyToPutAutoFollowRequest(p TuningParams, req *putautofollowpattern.Request) diag.Diagnostics {
	var diags diag.Diagnostics
	if v, d := OptIntFromInt64("max_outstanding_read_requests", p.MaxOutstandingReadRequests); d.HasError() {
		diags.Append(d...)
	} else if v != nil && *v > 0 {
		req.MaxOutstandingReadRequests = v
	}
	if v, d := OptIntFromInt64("max_outstanding_write_requests", p.MaxOutstandingWriteRequests); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxOutstandingWriteRequests = v
	}
	if v, d := OptIntFromInt64("max_read_request_operation_count", p.MaxReadRequestOperationCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxReadRequestOperationCount = v
	}
	if v := ByteSizeFromString(p.MaxReadRequestSize); v != nil {
		req.MaxReadRequestSize = v
	}
	if v := durationFromCustomDuration(p.MaxRetryDelay); v != nil {
		req.MaxRetryDelay = v
	}
	if v, d := OptIntFromInt64("max_write_buffer_count", p.MaxWriteBufferCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxWriteBufferCount = v
	}
	if v := ByteSizeFromString(p.MaxWriteBufferSize); v != nil {
		req.MaxWriteBufferSize = v
	}
	if v, d := OptIntFromInt64("max_write_request_operation_count", p.MaxWriteRequestOperationCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxWriteRequestOperationCount = v
	}
	if v := ByteSizeFromString(p.MaxWriteRequestSize); v != nil {
		req.MaxWriteRequestSize = v
	}
	if v := durationFromCustomDuration(p.ReadPollTimeout); v != nil {
		req.ReadPollTimeout = v
	}
	return diags
}

// ApplyToFollowRequest sets the tuning fields on req.
// MaxOutstandingReadRequests is *int64 in follow.Request; all other count fields are *int.
func ApplyToFollowRequest(p TuningParams, req *follow.Request) diag.Diagnostics {
	var diags diag.Diagnostics
	if v := typeutils.Int64Pointer(p.MaxOutstandingReadRequests); v != nil {
		req.MaxOutstandingReadRequests = v
	}
	if v, d := OptIntFromInt64("max_outstanding_write_requests", p.MaxOutstandingWriteRequests); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxOutstandingWriteRequests = v
	}
	if v, d := OptIntFromInt64("max_read_request_operation_count", p.MaxReadRequestOperationCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxReadRequestOperationCount = v
	}
	if v := ByteSizeFromString(p.MaxReadRequestSize); v != nil {
		req.MaxReadRequestSize = v
	}
	if v := durationFromCustomDuration(p.MaxRetryDelay); v != nil {
		req.MaxRetryDelay = v
	}
	if v, d := OptIntFromInt64("max_write_buffer_count", p.MaxWriteBufferCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxWriteBufferCount = v
	}
	if v := ByteSizeFromString(p.MaxWriteBufferSize); v != nil {
		req.MaxWriteBufferSize = v
	}
	if v, d := OptIntFromInt64("max_write_request_operation_count", p.MaxWriteRequestOperationCount); d.HasError() {
		diags.Append(d...)
	} else if v != nil {
		req.MaxWriteRequestOperationCount = v
	}
	if v := ByteSizeFromString(p.MaxWriteRequestSize); v != nil {
		req.MaxWriteRequestSize = v
	}
	if v := durationFromCustomDuration(p.ReadPollTimeout); v != nil {
		req.ReadPollTimeout = v
	}
	return diags
}

// ApplyToResumeFollowRequest sets the tuning fields on req.
// resumefollow.Request uses *int64 for all count fields and *string for byte sizes.
func ApplyToResumeFollowRequest(p TuningParams, req *resumefollow.Request) {
	if v := typeutils.Int64Pointer(p.MaxOutstandingReadRequests); v != nil {
		req.MaxOutstandingReadRequests = v
	}
	if v := typeutils.Int64Pointer(p.MaxOutstandingWriteRequests); v != nil {
		req.MaxOutstandingWriteRequests = v
	}
	if v := typeutils.Int64Pointer(p.MaxReadRequestOperationCount); v != nil {
		req.MaxReadRequestOperationCount = v
	}
	if v := typeutils.OptionalString(p.MaxReadRequestSize); v != nil {
		req.MaxReadRequestSize = v
	}
	if v := durationFromCustomDuration(p.MaxRetryDelay); v != nil {
		req.MaxRetryDelay = v
	}
	if v := typeutils.Int64Pointer(p.MaxWriteBufferCount); v != nil {
		req.MaxWriteBufferCount = v
	}
	if v := typeutils.OptionalString(p.MaxWriteBufferSize); v != nil {
		req.MaxWriteBufferSize = v
	}
	if v := typeutils.Int64Pointer(p.MaxWriteRequestOperationCount); v != nil {
		req.MaxWriteRequestOperationCount = v
	}
	if v := typeutils.OptionalString(p.MaxWriteRequestSize); v != nil {
		req.MaxWriteRequestSize = v
	}
	if v := durationFromCustomDuration(p.ReadPollTimeout); v != nil {
		req.ReadPollTimeout = v
	}
}

// durationFromCustomDuration returns an estypes.Duration when v is known, otherwise nil.
func durationFromCustomDuration(v customtypes.Duration) estypes.Duration {
	if !typeutils.IsKnown(v) {
		return nil
	}
	return estypes.Duration(v.ValueString())
}

// customDurationFromString converts a types.String to a customtypes.Duration,
// preserving null/unknown semantics.
func customDurationFromString(v types.String) customtypes.Duration {
	if v.IsUnknown() {
		return customtypes.NewDurationUnknown()
	}
	if v.IsNull() {
		return customtypes.NewDurationNull()
	}
	return customtypes.NewDurationValue(v.ValueString())
}

// TuningParamsFromParameters extracts TuningParams from an estypes.FollowerIndexParameters.
func TuningParamsFromParameters(params *estypes.FollowerIndexParameters) TuningParams {
	if params == nil {
		params = &estypes.FollowerIndexParameters{}
	}
	return TuningParams{
		MaxOutstandingReadRequests:    types.Int64PointerValue(params.MaxOutstandingReadRequests),
		MaxOutstandingWriteRequests:   typeutils.IntPointerToInt64Value(params.MaxOutstandingWriteRequests),
		MaxReadRequestOperationCount:  typeutils.IntPointerToInt64Value(params.MaxReadRequestOperationCount),
		MaxReadRequestSize:            ByteSizeToString(params.MaxReadRequestSize),
		MaxRetryDelay:                 customDurationFromString(typeutils.ElasticsearchDurationToString(params.MaxRetryDelay)),
		MaxWriteBufferCount:           typeutils.IntPointerToInt64Value(params.MaxWriteBufferCount),
		MaxWriteBufferSize:            ByteSizeToString(params.MaxWriteBufferSize),
		MaxWriteRequestOperationCount: typeutils.IntPointerToInt64Value(params.MaxWriteRequestOperationCount),
		MaxWriteRequestSize:           ByteSizeToString(params.MaxWriteRequestSize),
		ReadPollTimeout:               customDurationFromString(typeutils.ElasticsearchDurationToString(params.ReadPollTimeout)),
	}
}
