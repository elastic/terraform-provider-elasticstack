package clients

import (
	"context"

	"fixture/kbapi"
)

type Scoped struct {
	API *kbapi.ClientWithResponses
}

func (s *Scoped) Check(ctx context.Context) error {
	_, err := s.API.GetWidgetWithResponse(ctx, "probe")
	return err
}
