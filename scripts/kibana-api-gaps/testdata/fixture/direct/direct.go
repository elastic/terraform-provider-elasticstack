package direct

import (
	"context"

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
	_, err := c.CreateWidgetWithResponse(ctx, &kbapi.CreateWidgetParams{DryRun: &dry}, body)
	return err
}
