package copyassign

import (
	"context"

	"fixture/kbapi"
)

func fill(out, in *kbapi.Widget) {
	*out = *in
}

func Create(ctx context.Context, c *kbapi.ClientWithResponses, in *kbapi.Widget) error {
	var body kbapi.Widget
	fill(&body, in)
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
