package ctxwarp

import (
	"context"
)

type headerContextKey struct{}

type HeaderContext struct {
	RequestID     string `json:"request_id,omitempty"`
	CollectionKey string `json:"collection_key,omitempty"`
	UserKey       string `json:"user_key,omitempty"` // TODO: 需要删除, 后续通过上层服务将内容中的用户信息补全，此服务只做内容的提取索引.
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
