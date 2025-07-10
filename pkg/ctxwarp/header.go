package ctxwarp

import (
	"context"
)

type headerContextKey struct{}

type HeaderContext struct {
	RequestID    string `json:"request_id,omitempty"`
	CollectionID string `json:"collection_id,omitempty"`
	User         string `json:"user,omitempty"`
}

func SetHeaderContext(ctx context.Context, header *HeaderContext) context.Context {
	return context.WithValue(ctx, headerContextKey{}, header)
}

func GetHeaderContext(ctx context.Context) *HeaderContext {
	header := ctx.Value(headerContextKey{})
	if header == nil {
		return nil
	}
	return header.(*HeaderContext)
}
