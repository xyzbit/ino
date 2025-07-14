package konwledge

type KnowledgeExtractorResult struct {
	Source      string   `json:"source,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Facts       []string `json:"facts,omitempty"`
	Preferences []string `json:"preferences,omitempty"`
}
