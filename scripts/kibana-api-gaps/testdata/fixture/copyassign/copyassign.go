package copyassign

import (
	"context"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses, in *kbapi.Widget) error {
	var body kbapi.Widget
	body = *in
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
