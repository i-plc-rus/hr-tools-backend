package aichecker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hr-tools-backend/config"
	ollamamodels "hr-tools-backend/models/api/ollama"
	"net"
	"net/http"
	"time"
)

// IsTextAiAvailable проверяет доступность текстового AI
// Запускает промт и если ИИ занят, запрос оборвется через 15 сек и вернет false
func IsTextAiAvailable(ctx context.Context) (bool, error) {
	request := ollamamodels.OllamaRequest{
		Model:  "deepseek-r1:7b", // 1.7 для быстрых ответов
		Prompt: " ",
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return true, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, "POST", config.Conf.AI.Ollama.OllamaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return true, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return false, nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return false, nil
		}
		return true, err
	}
	defer resp.Body.Close()

	return true, nil
}
