package import_saved_objects

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/mapstructure"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	r.importObjects(ctx, request.Plan, &response.State, &response.Diagnostics)
}

func (r *Resource) importObjects(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	if !resourceReady(r, diags) {
		return
	}

	var model modelV0

	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return
	}

	kibanaClient, err := r.client.GetKibanaClient()
	if err != nil {
		diags.AddError("unable to get kibana client", err.Error())
		return
	}

	resp, err := kibanaClient.KibanaSavedObject.Import([]byte(model.FileContents.ValueString()), model.Overwrite.ValueBool(), model.SpaceID.ValueString())
	if err != nil {
		diags.AddError("failed to import saved objects", err.Error())
		return
	}

	var respModel responseModel

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &respModel,
		TagName: "json",
	})
	if err != nil {
		diags.AddError("failed to create model decoder", err.Error())
		return
	}

	err = decoder.Decode(resp)
	if err != nil {
		diags.AddError("failed to decode response", err.Error())
		return
	}

	if model.ID.IsUnknown() {
		model.ID = types.StringValue(uuid.NewString())
	}

	diags.Append(state.Set(ctx, model)...)
	diags.Append(state.SetAttribute(ctx, path.Root("success"), respModel.Success)...)
	diags.Append(state.SetAttribute(ctx, path.Root("success_count"), respModel.SuccessCount)...)
	diags.Append(state.SetAttribute(ctx, path.Root("errors"), respModel.Errors)...)
	diags.Append(state.SetAttribute(ctx, path.Root("success_results"), respModel.SuccessResults)...)
	if diags.HasError() {
		return
	}

	if !respModel.Success && !model.IgnoreImportErrors.ValueBool() {
		var detail strings.Builder
		for i, err := range respModel.Errors {
			detail.WriteString(fmt.Sprintf("import error [%d]: %s\n", i, err))
		}
		detail.WriteString("see the `errors` attribute for the full resposne")

		if respModel.SuccessCount > 0 {
			diags.AddWarning(
				"not all objects were imported successfully",
				detail.String(),
			)
		} else {
			diags.AddError(
				"no objects imported successfully",
				detail.String(),
			)
		}
	}
}

type responseModel struct {
	Success        bool            `json:"success"`
	SuccessCount   int             `json:"successCount"`
	Errors         []importError   `json:"errors"`
	SuccessResults []importSuccess `json:"successResults"`
}

type importSuccess struct {
	ID            string     `tfsdk:"id" json:"id"`
	Type          string     `tfsdk:"type" json:"type"`
	DestinationID string     `tfsdk:"destination_id" json:"destinationId"`
	Meta          importMeta `tfsdk:"meta" json:"meta"`
}

type importError struct {
	ID    string          `tfsdk:"id" json:"id"`
	Type  string          `tfsdk:"type" json:"type"`
	Title string          `tfsdk:"title" json:"title"`
	Error importErrorType `tfsdk:"error" json:"error"`
	Meta  importMeta      `tfsdk:"meta" json:"meta"`
}

func (ie importError) String() string {
	title := ie.Title
	if title == "" {
		title = ie.Meta.Title
	}

	return fmt.Sprintf("[%s] error on [%s] with ID [%s] and title [%s]", ie.Error.Type, ie.Type, ie.ID, title)
}

type importErrorType struct {
	Type string `tfsdk:"type" json:"type"`
}

type importMeta struct {
	Icon  string `tfsdk:"icon" json:"icon"`
	Title string `tfsdk:"title" json:"title"`
}
