package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tatsuo48/agito/internal/apperr"
)

// Config holds Ollama connection settings.
type Config struct {
	Host        string
	Model       string
	Temperature float64
	NumCtx      int
}

const systemPrompt = "You are a git assistant. You MUST write all output in English only. " +
	"Never use Japanese, Chinese, Korean, or any non-English language, " +
	"even if the diff you are analyzing contains non-English text."

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string         `json:"model"`
	Messages []chatMessage  `json:"messages"`
	Format   string         `json:"format"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options"`
}

type chatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Error string `json:"error,omitempty"`
}

// Generate calls Ollama /api/chat and returns the raw JSON string from the model.
func Generate(cfg Config, diffContent, extraInstruction string) (string, error) {
	userPrompt := BuildPrompt(diffContent, extraInstruction)
	body := chatRequest{
		Model: cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
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
	resp, err := client.Post(cfg.Host+"/api/chat", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("%w: %v (is ollama running?)", apperr.ErrOllama, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: reading response: %v", apperr.ErrOllama, err)
	}

	var result chatResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("%w: invalid response JSON: %v", apperr.ErrOllama, err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("%w: %s", apperr.ErrOllama, result.Error)
	}
	return result.Message.Content, nil
}
