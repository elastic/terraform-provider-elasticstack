package rawbody

import (
	"bytes"
	"context"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	verbose := true
	_, err := c.CreateWidgetWithBodyWithResponse(ctx, &kbapi.CreateWidgetParams{Verbose: &verbose}, "application/json", bytes.NewReader(nil))
	return err
}
