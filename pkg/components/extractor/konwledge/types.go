package konwledge

type KnowledgeExtractorResult struct {
	Facts       []string `json:"facts"`
	Preferences []string `json:"preferences"`
}
