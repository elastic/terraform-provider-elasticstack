package fxcore

import "context"

type Options[T any] struct {
	Create func(ctx context.Context, model T) error
}

type Resource[T any] struct {
	opts Options[T]
}

func New[T any](opts Options[T]) *Resource[T] {
	return &Resource[T]{opts: opts}
}

func (r *Resource[T]) Run(ctx context.Context, model T) error {
	return r.opts.Create(ctx, model)
}
