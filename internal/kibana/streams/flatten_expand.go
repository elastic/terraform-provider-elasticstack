package streams

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbstreams"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expandGroupToAPI converts the Terraform groupModel into the kbstreams payload used
// by PUT /api/streams/{name}/_group.
func expandGroupToAPI(ctx context.Context, m *groupModel) (*kbstreams.PutStreamsNameGroupJSONBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := &kbstreams.PutStreamsNameGroupJSONBody{}
	if m == nil {
		// Ensure we always send empty arrays/objects rather than nulls to match
		// the Streams Group schema expectations.
		body.Group.Members = []string{}
		body.Group.Metadata = map[string]string{}
		body.Group.Tags = []string{}
		return body, diags
	}

	// Start with empty values so we never send null for these fields.
	body.Group.Members = []string{}
	body.Group.Metadata = map[string]string{}
	body.Group.Tags = []string{}

	// members
	if len(m.Members) > 0 {
		members := make([]string, 0, len(m.Members))
		for _, v := range m.Members {
			if v.IsNull() || v.IsUnknown() {
				continue
			}
			members = append(members, v.ValueString())
		}
		body.Group.Members = members
	}

	// metadata
	if !m.Metadata.IsNull() && !m.Metadata.IsUnknown() {
		var meta map[string]string
		d := m.Metadata.ElementsAs(ctx, &meta, false)
		diags.Append(d...)
		if !diags.HasError() {
			body.Group.Metadata = meta
		}
	}

	// tags
	if len(m.Tags) > 0 {
		tags := make([]string, 0, len(m.Tags))
		for _, v := range m.Tags {
			if v.IsNull() || v.IsUnknown() {
				continue
			}
			tags = append(tags, v.ValueString())
		}
		body.Group.Tags = tags
	}

	return body, diags
}

// flattenGroupFromAPI populates a groupModel from the JSON returned by
// GET /api/streams/{name}/_group.
func flattenGroupFromAPI(ctx context.Context, apiBytes []byte, m *groupModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiBytes == nil {
		return diags
	}

	var body kbstreams.PutStreamsNameGroupJSONBody
	if err := json.Unmarshal(apiBytes, &body); err != nil {
		diags.AddError("Failed to decode group stream settings", err.Error())
		return diags
	}

	// members
	if len(body.Group.Members) == 0 {
		m.Members = nil
	} else {
		m.Members = make([]types.String, len(body.Group.Members))
		for i, v := range body.Group.Members {
			m.Members[i] = types.StringValue(v)
		}
	}

	// metadata
	if len(body.Group.Metadata) == 0 {
		// Preserve empty map rather than null to avoid post-apply inconsistencies
		// when the plan contained `metadata = {}`.
		mv, d := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.Metadata = mv
	} else {
		mv, d := types.MapValueFrom(ctx, types.StringType, body.Group.Metadata)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.Metadata = mv
	}

	// tags
	if len(body.Group.Tags) == 0 {
		// Preserve empty list rather than null to avoid post-apply inconsistencies
		// when the plan contained `tags = []`.
		m.Tags = []types.String{}
	} else {
		m.Tags = make([]types.String, len(body.Group.Tags))
		for i, v := range body.Group.Tags {
			m.Tags[i] = types.StringValue(v)
		}
	}

	return diags
}

// flattenIngestFromAPI decodes the JSON returned by GET /api/streams/{name}/_ingest
// into a Terraform Object value for the computed `ingest` attribute. For the
// current POC we only surface the ingest `type` field.
func flattenIngestFromAPI(ctx context.Context, apiBytes []byte) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiBytes == nil {
		return types.ObjectNull(ingestAttrTypes), diags
	}

	type ingestAPI struct {
		Ingest struct {
			Type string `json:"type"`
		} `json:"ingest"`
	}

	var body ingestAPI
	if err := json.Unmarshal(apiBytes, &body); err != nil {
		diags.AddError("Failed to decode ingest stream settings", err.Error())
		return types.ObjectNull(ingestAttrTypes), diags
	}

	attrs := map[string]attr.Value{
		"type": types.StringValue(body.Ingest.Type),
	}
	obj, d := types.ObjectValue(ingestAttrTypes, attrs)
	diags.Append(d...)

	return obj, diags
}

// expandGroupToStreamUpsertJSON builds the minimal JSON body required to create
// or upsert a group stream via PUT /api/streams/{name}. It mirrors the
// Streams.GroupStream.UpsertRequest shape by sending:
//   - dashboards, rules, queries as empty arrays (emptyAssets)
//   - stream.description (when set)
//   - stream.group.metadata, stream.group.tags, stream.group.members from the plan.
func expandGroupToStreamUpsertJSON(ctx context.Context, name string, plan *streamModel) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The stream name is taken from the URL path in the Streams API and is
	// intentionally omitted from the JSON body. The server-side schema expects
	// `stream.name` to be undefined in the payload.
	// For the group-stream upsert branch, `stream.description` is required,
	// while `stream.ingest` must not be present (it is validated separately
	// via the _ingest endpoint for ingest streams).
	desc := ""
	if plan != nil && !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		desc = plan.Description.ValueString()
	}

	stream := map[string]any{
		"description": desc,
	}

	// Group block
	if plan != nil && plan.Group != nil {
		group := map[string]any{}

		// metadata
		if !plan.Group.Metadata.IsNull() && !plan.Group.Metadata.IsUnknown() {
			var meta map[string]string
			d := plan.Group.Metadata.ElementsAs(ctx, &meta, false)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			group["metadata"] = meta
		} else {
			group["metadata"] = map[string]string{}
		}

		// tags
		if len(plan.Group.Tags) > 0 {
			tags := make([]string, 0, len(plan.Group.Tags))
			for _, v := range plan.Group.Tags {
				if v.IsNull() || v.IsUnknown() {
					continue
				}
				tags = append(tags, v.ValueString())
			}
			group["tags"] = tags
		} else {
			group["tags"] = []string{}
		}

		// members
		if len(plan.Group.Members) > 0 {
			members := make([]string, 0, len(plan.Group.Members))
			for _, v := range plan.Group.Members {
				if v.IsNull() || v.IsUnknown() {
					continue
				}
				members = append(members, v.ValueString())
			}
			group["members"] = members
		} else {
			// The schema allows an empty members array; we preserve that here.
			group["members"] = []string{}
		}

		stream["group"] = group
	}

	body := map[string]any{
		"dashboards": []string{},
		"rules":      []string{},
		"queries":    []any{},
		"stream":     stream,
	}

	raw, err := json.Marshal(body)
	if err != nil {
		diags.AddError("Failed to encode group stream upsert payload", err.Error())
		return nil, diags
	}

	return raw, diags
}
