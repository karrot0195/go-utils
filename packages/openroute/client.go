package openroute

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	OpenRouterAPI = "https://openrouter.ai"
	ModelPath     = "/api/v1/models"
)

// FetchModels downloads the model list from OpenRouter, pretty-prints it,
// and writes it to a dated cache file. It returns the raw JSON bytes.
//
// If a cache file for today already exists, it is returned instead of
// performing a network call.
func FetchModels() ([]byte, string, error) {
	cacheFile := os.TempDir() + "/" + time.Now().Format("020106") + "-model_list.json"

	// Use cached file if it exists
	if data, err := os.ReadFile(cacheFile); err == nil {
		return data, cacheFile, nil
	}

	resp, err := http.Get(OpenRouterAPI + ModelPath)
	if err != nil {
		return nil, "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	// Pretty-print JSON for readability before saving
	var parsed any
	if err := json.Unmarshal(body, &parsed); err == nil {
		if formatted, err := json.MarshalIndent(parsed, "", "  "); err == nil {
			body = formatted
		}
	}

	if err := os.WriteFile(cacheFile, body, 0644); err != nil {
		return nil, "", fmt.Errorf("write cache: %w", err)
	}

	return body, cacheFile, nil
}

// ParseModels unmarshals raw JSON into a ModelResponse.
func ParseModels(data []byte) (*ModelResponse, error) {
	var resp ModelResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse model list: %w", err)
	}
	return &resp, nil
}

// FilterModels returns models matching the given type.
// Supported types: "free" (id ends with ":free") and "all".
func FilterModels(models []Model, modelType string) []Model {
	out := make([]Model, 0, len(models))
	for _, m := range models {
		if modelType == "all" || hasFreeSuffix(m.ID) {
			out = append(out, m)
		}
	}
	return out
}

func hasFreeSuffix(id string) bool {
	const suffix = ":free"
	return len(id) >= len(suffix) && id[len(id)-len(suffix):] == suffix
}
