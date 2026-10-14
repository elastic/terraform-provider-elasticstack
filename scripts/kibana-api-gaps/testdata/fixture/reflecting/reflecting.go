package reflecting

import (
	"context"
	"reflect"

	"fixture/kbapi"
)

func fill[T any](props *T, name string) {
	reflect.ValueOf(props).Elem().FieldByName("Name").SetString(name)
}

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	var body kbapi.CreateWidgetJSONRequestBody
	fill(&body, "w")
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
