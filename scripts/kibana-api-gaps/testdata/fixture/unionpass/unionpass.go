package unionpass

import (
	"context"
	"encoding/json"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses, raw string) error {
	var body kbapi.CreateBlobJSONRequestBody
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return err
	}
	_, err := c.CreateBlobWithResponse(ctx, body)
	return err
}
