package reflectwrap

import (
	"context"
	"reflect"

	"fixture/kbapi"
)

func outer(x any) { inner(x) }

func inner(x any) { reflect.ValueOf(x).Elem().FieldByName("Name").SetString("w") }

type filler interface{ fill(v any) }

type reflectFiller struct{}

func (reflectFiller) fill(v any) { reflect.ValueOf(v).Elem().FieldByName("Mode") }

func Create(ctx context.Context, c *kbapi.ClientWithResponses) error {
	var body kbapi.CreateWidgetJSONRequestBody
	outer(&body)
	var settings kbapi.Settings
	var f filler = reflectFiller{}
	f.fill(&settings)
	_, err := c.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
