package noop

import (
	"context"
)

type keyType string

const noopKey keyType = "noop_key"

func ContainsNoop(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	v := ctx.Value(noopKey)
	if v == nil {
		return false
	}
	return v.(bool)
}

func NewCtxWithNoop(ctx context.Context, isNoop bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if ContainsNoop(ctx) { // todo needed?
		return ctx
	}
	noopCtx := context.WithValue(ctx, noopKey, isNoop)
	return noopCtx
}
