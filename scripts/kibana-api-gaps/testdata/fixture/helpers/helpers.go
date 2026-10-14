package helpers

import (
	"context"

	"fixture/kbapi"
)

type Writer interface {
	Describe() *string
}

func Create(ctx context.Context, c *kbapi.ClientWithResponses, w Writer) error {
	body := kbapi.CreateWidgetJSONRequestBody{Name: "w", Description: w.Describe()}
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}

type Legacy struct{}

func (Legacy) Describe() *string {
	return nil
}

func Read(ctx context.Context, c *kbapi.ClientWithResponses, id string) error {
	_, err := c.GetWidgetWithResponse(ctx, id)
	return err
}
