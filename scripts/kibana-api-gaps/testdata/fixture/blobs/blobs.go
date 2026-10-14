package blobs

import (
	"context"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	text := "t"
	var body kbapi.CreateBlobJSONRequestBody
	var variant kbapi.CreateBlobJSONBody
	_ = variant
	_ = text
	_, err := c.CreateBlobWithResponse(ctx, body)
	return err
}
