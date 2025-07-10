package konwledge

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
)

type KnowledgeExtractor struct {
}

func NewKnowledgeExtractor(ctx context.Context, prompt *prompt.Prompt) (*KnowledgeExtractor, error) {
	return &KnowledgeExtractor{prompt: PromptKnowledgeExtractor}, nil
}
