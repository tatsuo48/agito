package generate

import (
	"fmt"

	"github.com/moneyforward/figaro/internal/ollama"
)

const maxRetries = 2

// Run calls Ollama, validates the output, and retries up to maxRetries times on validation failure.
func Run(cfg ollama.Config, diffContent, language, extraInstruction string) (*Content, error) {
	var lastErr error
	prevInvalid := ""

	for attempt := 0; attempt <= maxRetries; attempt++ {
		extra := extraInstruction
		if prevInvalid != "" {
			if extra != "" {
				extra += "\n"
			}
			extra += "Previous output was invalid because: " + prevInvalid
		}

		raw, err := ollama.Generate(cfg, diffContent, language, extra)
		if err != nil {
			return nil, err
		}

		content, err := ParseAndValidate([]byte(raw))
		if err != nil {
			prevInvalid = err.Error()
			lastErr = fmt.Errorf("validation failed (attempt %d): %w", attempt+1, err)
			continue
		}
		return content, nil
	}
	return nil, lastErr
}
