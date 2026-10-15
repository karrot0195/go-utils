package openroute

import (
	"fmt"
	"sort"
	"strconv"
)

// PrintModels filters, sorts, and prints the model list in a tabular format.
func PrintModels(resp *ModelResponse, modelType string) {
	free := FilterModels(resp.Data, modelType)

	sort.Slice(free, func(i, j int) bool {
		return score(free[i]) > score(free[j])
	})

	fmt.Printf("%-55s %-55s %-15s %-15s %-15s %-15s %-15s %s\n",
		"ID", "NAME", "CODING_INDEX", "INTELL_INDEX", "SUM", "PROMPT", "COMPLETION", "CONTEXT")

	for _, m := range free {
		coding, intel, sum := "-", "-", "-"
		prompt, completion := "-", "-"

		if m.Benchmarks.ArtificialAnalysis != nil {
			b := m.Benchmarks.ArtificialAnalysis
			coding = fmt.Sprintf("%.2f", b.CodingIndex)
			intel = fmt.Sprintf("%.2f", b.IntelligenceIndex)
			sum = fmt.Sprintf("%.2f", b.CodingIndex+b.IntelligenceIndex)
		}
		prompt = formatPrice(m.Price.Prompt)
		completion = formatPrice(m.Price.Completion)

		fmt.Printf("%-55s %-55s %-15s %-15s %-15s %-15s %-15s %d\n",
			m.ID, m.Name, coding, intel, sum, prompt, completion, m.ContextLength)
	}
}

// score is the sort key: coding_index + intelligence_index (0 if nil).
func score(m Model) float64 {
	if m.Benchmarks.ArtificialAnalysis == nil {
		return 0
	}
	b := m.Benchmarks.ArtificialAnalysis
	return b.CodingIndex + b.IntelligenceIndex
}

// formatPrice parses an OpenRouter price string (per-token cost) and scales
// it to per-million-token cost for readability.
func formatPrice(s string) string {
	if s == "" {
		return "-"
	}
	p, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%.4f", p*1e6)
}
