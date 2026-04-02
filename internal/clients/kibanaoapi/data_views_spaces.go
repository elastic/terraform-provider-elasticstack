package kibanaoapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type updateObjectsSpacesRequest struct {
	Objects        []savedObjectRef `json:"objects"`
	SpacesToAdd    []string         `json:"spacesToAdd"`
	SpacesToRemove []string         `json:"spacesToRemove"`
}

type savedObjectRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// UpdateDataViewNamespaces diffs old/new namespaces and calls
func UpdateDataViewNamespaces(
	ctx context.Context,
	client *Client,
	spaceID string,
	dataViewID string,
	oldNamespaces []string,
	newNamespaces []string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	spacesToAdd := []string{}
	spacesToRemove := []string{}

	for _, ns := range newNamespaces {
		if !slices.Contains(oldNamespaces, ns) {
			spacesToAdd = append(spacesToAdd, ns)
		}
	}
	for _, ns := range oldNamespaces {
		if !slices.Contains(newNamespaces, ns) {
			spacesToRemove = append(spacesToRemove, ns)
		}
	}

	if len(spacesToAdd) == 0 && len(spacesToRemove) == 0 {
		return diags
	}

	reqBody := updateObjectsSpacesRequest{
		Objects:        []savedObjectRef{{Type: "index-pattern", ID: dataViewID}},
		SpacesToAdd:    spacesToAdd,
		SpacesToRemove: spacesToRemove,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		diags.AddError("Failed to marshal spaces request", err.Error())
		return diags
	}

	path := BuildSpaceAwarePath(spaceID, "/api/spaces/_update_objects_spaces")
	url := client.URL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		diags.AddError("Failed to create spaces request", err.Error())
		return diags
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTP.Do(req)
	if err != nil {
		diags.AddError("Failed to update data view namespaces", err.Error())
		return diags
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		diags.AddError(
			fmt.Sprintf("Unexpected status %d when updating data view namespaces", resp.StatusCode),
			fmt.Sprintf("dataViewID=%s spacesToAdd=%v spacesToRemove=%v response=%s",
				dataViewID, spacesToAdd, spacesToRemove, string(body)),
		)
	}

	return diags
}
