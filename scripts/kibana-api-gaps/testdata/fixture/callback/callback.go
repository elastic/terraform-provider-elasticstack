package callback

import (
	"context"

	"fixture/fxcore"
	"fixture/kbapi"
)

type model struct {
	client *kbapi.ClientWithResponses
}

func New() *fxcore.Resource[model] {
	return fxcore.New(fxcore.Options[model]{Create: create})
}

func create(ctx context.Context, m model) error {
	desc := "d"
	retries := 3
	body := kbapi.CreateWidgetJSONRequestBody{Name: "w", Description: &desc}
	body.Settings = &kbapi.Settings{Retries: &retries}
	var source kbapi.Source
	if err := source.FromSourceA(kbapi.SourceA{A: &desc}); err != nil {
		return err
	}
	body.Source = &source
	_, err := m.client.CreateWidgetWithResponse(ctx, nil, body)
	return err
}
