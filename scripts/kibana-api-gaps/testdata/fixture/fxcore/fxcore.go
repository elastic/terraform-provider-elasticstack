package fxcore

import (
	"context"

	"fixture/kbapi"
)

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

type Stepper interface {
	Step(ctx context.Context) error
}

type Runner[T any] struct {
	client *kbapi.ClientWithResponses
}

func (r *Runner[T]) Step(ctx context.Context) error {
	_, err := r.client.GetWidgetWithResponse(ctx, "step")
	return err
}

type Envelope[T any] struct {
	client *kbapi.ClientWithResponses
}

// Handle is invoked by the framework, never by name from a resource package.
func (e *Envelope[T]) Handle(ctx context.Context) error {
	_, err := e.client.CreateWidgetWithResponse(ctx, nil, kbapi.Widget{Name: "n"})
	return err
}
