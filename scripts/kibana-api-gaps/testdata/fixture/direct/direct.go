package direct

import (
	"context"
	"encoding/json"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	mode := "fast"
	dry := true
	body := kbapi.CreateWidgetJSONRequestBody{
		Name:     "w",
		Labels:   &[]kbapi.Label{{Key: "k"}},
		Settings: &kbapi.Settings{Mode: &mode},
	}
	if _, err := json.Marshal(body); err != nil {
		return err
	}
	_, err := c.CreateWidgetWithResponse(ctx, &kbapi.CreateWidgetParams{DryRun: &dry}, body)
	return err
}
