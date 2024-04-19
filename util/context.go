package util

import "context"

func ContextFromDict(pairs map[any]any) context.Context {
	ctx := context.Background()
	for key, val := range pairs {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}
