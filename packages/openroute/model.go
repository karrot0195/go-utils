package openroute

// Model mirrors the relevant fields from the OpenRouter /models response.
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       struct {
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
	} `json:"pricing"`
	ContextLength int `json:"context_length"`
	Benchmarks    struct {
		ArtificialAnalysis *struct {
			AgenticIndex      float64 `json:"agentic_index"`
			CodingIndex       float64 `json:"coding_index"`
			IntelligenceIndex float64 `json:"intelligence_index"`
		} `json:"artificial_analysis"`
	} `json:"benchmarks"`
}

// ModelResponse is the top-level JSON envelope returned by OpenRouter.
type ModelResponse struct {
	Data []Model `json:"data"`
}