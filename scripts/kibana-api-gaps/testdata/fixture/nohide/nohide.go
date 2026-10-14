package nohide

import (
	"context"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	mode := "fast"
	first := kbapi.Label{Key: "k"}
	second := first
	settings := kbapi.Settings(kbapi.Settings{Mode: &mode})
	body := kbapi.Widget{Name: "n", Labels: &[]kbapi.Label{second}, Settings: &settings}
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
