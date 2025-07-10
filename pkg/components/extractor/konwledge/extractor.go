package konwledge

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/xyzbit/ino/pkg/constants"
)

const (
	typ = "knowledge_extractor"
)

type Extractor struct {
	config *Config
}

type Config struct {
	// Extractor is the model used to extract facts and peferences from documents
	// Required
	Extractor model.ToolCallingChatModel
}

func NewExtractor(ctx context.Context, conf *Config) (*Extractor, error) {
	// conf check
	if err := conf.check(); err != nil {
		return nil, err
	}

	return &Extractor{
		config: conf,
	}, nil
}

func (i *Extractor) GetType() string {
	return typ
}

func (i *Extractor) IsCallbacksEnabled() bool {
	return true
}
func (k *Extractor) Extract(ctx context.Context, docs []*schema.Document, opts ...any) (results []*schema.Document, err error) {
	ctx = callbacks.EnsureRunInfo(ctx, k.GetType(), components.ComponentOfIndexer)
	// callback info on start
	ctx = callbacks.OnStart(ctx, &indexer.CallbackInput{
		Docs: docs,
	})
	defer func() {
		if err != nil {
			callbacks.OnError(ctx, err)
		}
	}()

	if len(docs) == 0 {
		return []*schema.Document{}, nil
	}

	results = make([]*schema.Document, 0, len(docs))
	allIDs := make([]string, 0, len(docs))
	for _, doc := range docs {
		msgs, err := PromptKnowledgeExtractor.Format(ctx, map[string]any{
			"time_now":       time.Now().Format(time.DateTime),
			"origin_request": doc.Content,
		})
		if err != nil {
			return nil, err
		}

		output, err := k.config.Extractor.Generate(ctx, msgs)
		if err != nil {
			return nil, err
		}

		var result KnowledgeExtractorResult
		if err := json.Unmarshal([]byte(output.Content), &result); err != nil {
			return nil, err
		}

		for _, fact := range result.Facts {
			results = append(results, &schema.Document{
				Content: fact,
				MetaData: map[string]any{
					constants.CategoryKey: constants.CategoryValueFact,
				},
			})
		}

		for _, preference := range result.Preferences {
			results = append(results, &schema.Document{
				Content: preference,
				MetaData: map[string]any{
					constants.CategoryKey: constants.CategoryValuePreference,
				},
			})
		}

		allIDs = append(allIDs, doc.ID)
	}

	callbacks.OnEnd(ctx, &indexer.CallbackOutput{
		IDs: allIDs,
	})

	return results, nil
}

func (c *Config) check() error {
	// check extractor
	if c.Extractor == nil {
		return fmt.Errorf("extractor is required")
	}

	return nil
}
