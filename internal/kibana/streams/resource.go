package streams

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbstreams"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const resourceName = "kibana_stream"

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// Resource implements the elasticstack_kibana_stream resource.
type Resource struct {
	client *clients.ApiClient
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, resourceName)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// We use the standard composite ID format: <cluster_uuid>/<stream_name>.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "Expected configured API client, but got nil.")
		return
	}

	var plan streamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Name.IsUnknown() || plan.Name.IsNull() {
		resp.Diagnostics.AddError("Missing stream name", "The 'name' attribute must be set for kibana_stream.")
		return
	}

	// Determine which mode this resource is using.
	// In the current POC:
	//   - If a `group` block is present, we manage the group stream settings.
	//   - If no `group` block is present, we assume this is an ingest-only
	//     stream and treat it as a read-only view of `/api/streams/{name}/_ingest`.
	isGroupStream := plan.Group != nil
	isIngestStream := !isGroupStream

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana OAPI client", err.Error())
		return
	}

	name := plan.Name.ValueString()
	if isGroupStream {
		// Group stream mode: manage /_group configuration.
		tflog.Debug(ctx, "Creating Kibana group stream", map[string]any{
			"stream_name": name,
		})

		// If requested, create the base group stream via PUT /api/streams/{name}
		// when it does not already exist. This keeps the default behaviour
		// "attach to existing" unless create_if_missing is explicitly enabled.
		if !plan.CreateIfMissing.IsNull() && plan.CreateIfMissing.ValueBool() {
			tflog.Debug(ctx, "Ensuring base group stream exists via PUT /api/streams/{name}", map[string]any{
				"stream_name": name,
			})

			existingJSON, d := kibana_oapi.GetStreamJSON(ctx, client, name)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			if existingJSON == nil {
				rawBody, d := expandGroupToStreamUpsertJSON(ctx, name, &plan)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}

				d = kibana_oapi.PutStreamRaw(ctx, client, name, rawBody)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		groupBody, d := expandGroupToAPI(ctx, plan.Group)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Upsert group configuration for the stream.
		apiBody := kbstreams.PutStreamsNameGroupJSONRequestBody(*groupBody)
		d = kibana_oapi.PutStreamGroup(ctx, client, name, apiBody)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Read back the group configuration to populate state deterministically.
		groupJSON, d := kibana_oapi.GetStreamGroupJSON(ctx, client, name)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if groupJSON == nil {
			// Treat missing group as the resource having been removed.
			resp.State.RemoveResource(ctx)
			return
		}

		if plan.Group == nil {
			plan.Group = &groupModel{}
		}
		d = flattenGroupFromAPI(ctx, groupJSON, plan.Group)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Group-only streams have no ingest settings; ensure the computed
		// `ingest` attribute is set to a known null object after apply.
		plan.Ingest = types.ObjectNull(ingestAttrTypes)
	} else if isIngestStream {
		// Ingest stream mode (POC, read-only): do not attempt to create or update
		// ingest via Terraform yet. Instead, validate the ingest stream exists and
		// hydrate state from the existing _ingest definition.
		tflog.Debug(ctx, "Registering existing Kibana ingest stream", map[string]any{
			"stream_name": name,
		})

		ingestJSON, d := kibana_oapi.GetStreamIngestJSON(ctx, client, name)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if ingestJSON == nil {
			resp.Diagnostics.AddError(
				"Ingest stream not found",
				fmt.Sprintf("No ingest definition was found for stream %q. Ensure the ingest stream exists in Kibana before managing it with Terraform.", name),
			)
			return
		}

		plan.Ingest, d = flattenIngestFromAPI(ctx, ingestJSON)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Compute the composite ID <cluster_uuid>/<name>.
	var fwDiags fwdiag.Diagnostics
	compID, sdkDiags := r.client.ID(ctx, name)
	fwDiags = diagutil.FrameworkDiagsFromSDK(sdkDiags)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if compID != nil {
		plan.ID = types.StringValue(compID.String())
	} else {
		plan.ID = types.StringValue(name)
	}

	// For now, we distinguish only between "group" and ingest-type streams.
	if isGroupStream {
		plan.Type = types.StringValue("group")
	} else if isIngestStream {
		// For ingest streams in this POC we currently surface only that this is an
		// ingest stream; more detailed typing (wired/classic) can be added later.
		plan.Type = types.StringValue("ingest")
	}

	// space_id is currently a placeholder; we default to "default" while Streams
	// is global in the Streams OAS snapshot.
	if plan.SpaceID.IsNull() || plan.SpaceID.IsUnknown() {
		plan.SpaceID = types.StringValue("default")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "Expected configured API client, but got nil.")
		return
	}

	var state streamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana OAPI client", err.Error())
		return
	}

	// Derive the logical stream name either from the composite ID or from the
	// stored name attribute.
	var name string
	if !state.ID.IsNull() && state.ID.ValueString() != "" {
		compID, fwDiags := clients.CompositeIdFromStrFw(state.ID.ValueString())
		resp.Diagnostics.Append(fwDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if compID != nil {
			name = compID.ResourceId
		}
	}
	if name == "" && !state.Name.IsNull() && !state.Name.IsUnknown() {
		name = state.Name.ValueString()
	}
	if name == "" {
		resp.Diagnostics.AddError("Missing stream identifier", "Neither 'id' nor 'name' could be resolved from state.")
		return
	}

	// Decide which aspects of the stream to refresh. This allows us to support
	// group-only, ingest-only, or combined views without calling endpoints that
	// are invalid for a given stream type.
	hasGroupState := state.Group != nil

	if hasGroupState {
		tflog.Debug(ctx, "Reading Kibana group stream", map[string]any{
			"stream_name": name,
		})

		groupJSON, d := kibana_oapi.GetStreamGroupJSON(ctx, client, name)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		if groupJSON == nil {
			// Stream (or its group configuration) no longer exists.
			resp.State.RemoveResource(ctx)
			return
		}

		if state.Group == nil {
			state.Group = &groupModel{}
		}
		d = flattenGroupFromAPI(ctx, groupJSON, state.Group)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Best-effort read of ingest settings for this stream. For the current POC we
	// treat ingest as read-only configuration: we reflect whatever Kibana
	// returns, but Create/Update do not yet push ingest changes.
	ingestJSON, d := kibana_oapi.GetStreamIngestJSON(ctx, client, name)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Ingest, d = flattenIngestFromAPI(ctx, ingestJSON)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep ID stable; recompute if it was missing.
	if state.ID.IsNull() || state.ID.ValueString() == "" {
		compID, sdkDiags := r.client.ID(ctx, name)
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if compID != nil {
			state.ID = types.StringValue(compID.String())
		} else {
			state.ID = types.StringValue(name)
		}
	}

	state.Name = types.StringValue(name)
	state.Type = types.StringValue("group")
	if state.SpaceID.IsNull() || state.SpaceID.IsUnknown() {
		state.SpaceID = types.StringValue("default")
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "Expected configured API client, but got nil.")
		return
	}

	var plan streamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isGroupStream := plan.Group != nil
	isIngestStream := !isGroupStream

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana OAPI client", err.Error())
		return
	}

	if plan.Name.IsUnknown() || plan.Name.IsNull() {
		resp.Diagnostics.AddError("Missing stream name", "The 'name' attribute must be set for kibana_stream.")
		return
	}
	name := plan.Name.ValueString()

	if isGroupStream {
		tflog.Debug(ctx, "Updating Kibana group stream", map[string]any{
			"stream_name": name,
		})

		groupBody, d := expandGroupToAPI(ctx, plan.Group)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		apiBody := kbstreams.PutStreamsNameGroupJSONRequestBody(*groupBody)
		d = kibana_oapi.PutStreamGroup(ctx, client, name, apiBody)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Re-read after update to keep state in sync.
		groupJSON, d := kibana_oapi.GetStreamGroupJSON(ctx, client, name)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if groupJSON == nil {
			resp.State.RemoveResource(ctx)
			return
		}

		if plan.Group == nil {
			plan.Group = &groupModel{}
		}
		d = flattenGroupFromAPI(ctx, groupJSON, plan.Group)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Group-only streams have no ingest settings; ensure the computed
		// `ingest` attribute is set to a known null object after apply.
		plan.Ingest = types.ObjectNull(ingestAttrTypes)
	} else if isIngestStream {
		// Ingest streams are currently read-only from Terraform's perspective.
		// We refresh state from the existing _ingest definition but do not push
		// any changes to Kibana.
		tflog.Debug(ctx, "Refreshing existing Kibana ingest stream (read-only)", map[string]any{
			"stream_name": name,
		})

		ingestJSON, d := kibana_oapi.GetStreamIngestJSON(ctx, client, name)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if ingestJSON == nil {
			resp.State.RemoveResource(ctx)
			return
		}

		plan.Ingest, d = flattenIngestFromAPI(ctx, ingestJSON)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Preserve or compute ID as in Create.
	if plan.ID.IsNull() || plan.ID.ValueString() == "" {
		compID, sdkDiags := r.client.ID(ctx, name)
		resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if compID != nil {
			plan.ID = types.StringValue(compID.String())
		} else {
			plan.ID = types.StringValue(name)
		}
	}

	if isGroupStream {
		plan.Type = types.StringValue("group")
	} else if isIngestStream {
		// For ingest streams in this POC we currently surface only that this is an
		// ingest stream; more detailed typing (wired/classic) can be added later.
		plan.Type = types.StringValue("ingest")
	}
	if plan.SpaceID.IsNull() || plan.SpaceID.IsUnknown() {
		plan.SpaceID = types.StringValue("default")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "Expected configured API client, but got nil.")
		return
	}

	var state streamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Decide whether this resource represents a group stream or an ingest
	// stream. For the current POC we only issue a DELETE to Kibana for group
	// streams; ingest-only streams are treated as read-only views and are
	// simply removed from Terraform state.
	isGroupStream := state.Group != nil
	isIngestStream := !isGroupStream

	client, err := r.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kibana OAPI client", err.Error())
		return
	}

	// Derive the stream name from ID or name.
	var name string
	if !state.ID.IsNull() && state.ID.ValueString() != "" {
		compID, fwDiags := clients.CompositeIdFromStrFw(state.ID.ValueString())
		resp.Diagnostics.Append(fwDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if compID != nil {
			name = compID.ResourceId
		}
	}
	if name == "" && !state.Name.IsNull() && !state.Name.IsUnknown() {
		name = state.Name.ValueString()
	}
	if name == "" {
		resp.Diagnostics.AddError("Missing stream identifier", "Neither 'id' nor 'name' could be resolved from state.")
		return
	}

	// Ingest-only streams: do not attempt to delete anything in Kibana in this
	// POC. We only drop them from Terraform state.
	if isIngestStream && !isGroupStream {
		tflog.Debug(ctx, "Removing ingest-only stream from Terraform state without deleting in Kibana (read-only ingest POC)", map[string]any{
			"stream_name": name,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "Deleting Kibana stream", map[string]any{
		"stream_name": name,
	})

	d := kibana_oapi.DeleteStream(ctx, client, name)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
