package generic

import (
	"context"

	"fixture/fxcore"
)

func Run(ctx context.Context, s fxcore.Stepper) error {
	return s.Step(ctx)
}
