package helperuser

import (
	"context"

	"fixture/helpers"
	"fixture/kbapi"
)

func Run(ctx context.Context, c *kbapi.ClientWithResponses) error {
	return helpers.Create(ctx, c, helpers.Legacy{})
}
