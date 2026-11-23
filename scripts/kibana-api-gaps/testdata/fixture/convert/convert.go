package convert

import (
	"context"

	"fixture/kbapi"
)

type mine kbapi.Widget

func Create(ctx context.Context, c *kbapi.ClientWithResponses, m mine) error {
	body := kbapi.Widget(m)
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
