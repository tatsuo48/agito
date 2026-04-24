package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/moneyforward/figaro/internal/apperr"
)

// Config holds Ollama connection settings.
type Config struct {
	Host        string
	Model       string
	Temperature float64
	NumCtx      int
}

type generateRequest struct {
	Model   string         `json:"model"`
	System  string         `json:"system"`
	Prompt  string         `json:"prompt"`
	Format  string         `json:"format"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options"`
}

const systemPrompt = "You are a git assistant. You MUST respond only in English. Never use Japanese, Chinese, Korean, or any language other than English in your output, regardless of the language in the diff you are analyzing."

type generateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// Generate calls Ollama and returns the raw JSON string from the model.
func Generate(cfg Config, diffContent, extraInstruction string) (string, error) {
	prompt := BuildPrompt(diffContent, extraInstruction)
	body := generateRequest{
		Model:  cfg.Model,
		System: systemPrompt,
		Prompt: prompt,
		Format: "json",
		Stream: false,
		Options: map[string]any{
			"temperature": cfg.Temperature,
			"num_ctx":     cfg.NumCtx,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(cfg.Host+"/api/generate", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("%w: %v (is ollama running?)", apperr.ErrOllama, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: reading response: %v", apperr.ErrOllama, err)
	}

	var result generateResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("%w: invalid response JSON: %v", apperr.ErrOllama, err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("%w: %s", apperr.ErrOllama, result.Error)
	}
	return result.Response, nil
}
