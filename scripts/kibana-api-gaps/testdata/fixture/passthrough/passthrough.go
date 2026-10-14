package passthrough

import (
	"context"
	"encoding/json"

	"fixture/kbapi"
)

func Create(ctx context.Context, c *kbapi.ClientWithResponses, raw string) error {
	var body kbapi.CreateWidgetJSONRequestBody
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return err
	}
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
