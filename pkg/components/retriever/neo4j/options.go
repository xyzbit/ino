package neo4j

import (
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
)

// RetrieverImplOptions 检索器实现选项
type RetrieverImplOptions struct {
	TopK                int
	Limit               int
	SimilarityThreshold float64
	MaxDepth            int
	Filter              map[string]string
}

func WithTopK(topK int) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverImplOptions) {
		o.TopK = topK
	})
}

func WithLimit(limit int) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverImplOptions) {
		o.Limit = limit
	})
}

func WithSimilarityThreshold(similarityThreshold float64) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverImplOptions) {
		o.SimilarityThreshold = similarityThreshold
	})
}

func WithMaxDepth(maxDepth int) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverImplOptions) {
		o.MaxDepth = maxDepth
	})
}

// {user_key: "123", collection_key: "456"}
func WithFilter(filterKVs map[string]string) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverImplOptions) {
		o.Filter = filterKVs
	})
}

func (r *RetrieverImplOptions) GetFilter() string {
	if len(r.Filter) == 0 {
		return ""
	}
	filter := ""
	for k, v := range r.Filter {
		filter += fmt.Sprintf(" AND node.%s = \"%s\"", k, v)
	}
	return filter
}
