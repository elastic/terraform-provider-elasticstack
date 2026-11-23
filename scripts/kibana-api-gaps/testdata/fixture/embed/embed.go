package embed

import "fixture/fxcore"

type model struct{}

type Resource struct {
	*fxcore.Envelope[model]
}
