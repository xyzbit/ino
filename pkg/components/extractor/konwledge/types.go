package konwledge

type KnowledgeExtractorResult struct {
	Source      string   `json:"source"`
	Facts       []string `json:"facts"`
	Preferences []string `json:"preferences"`
}
