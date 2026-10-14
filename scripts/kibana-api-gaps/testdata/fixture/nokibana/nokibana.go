package nokibana

import "context"

type checker interface {
	Check(ctx context.Context) error
}

func Verify(ctx context.Context, c checker) error {
	return c.Check(ctx)
}

func Name() string {
	return "elasticsearch only"
}
