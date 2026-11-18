package reflectany

import (
	"context"
	"reflect"

	"fixture/kbapi"
)

func fillAny(props any) {
	reflect.ValueOf(props).Elem().FieldByName("Name").SetString("w")
}

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	var body kbapi.CreateWidgetJSONRequestBody
	fillAny(&body)
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}

func direct() {
	var settings kbapi.Settings
	_ = reflect.TypeOf(settings)
}
