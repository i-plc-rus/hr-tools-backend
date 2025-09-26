package ollamamodels

import "hr-tools-backend/config"

// Структуры для работы с Ollama API
type OllamaRequest struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Stream  bool    `json:"stream"`
	Options Options `json:"options"`
}

type Options struct {
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	NumPredict    int      `json:"num_predict,omitempty"` // Аналог MaxTokens
	RepeatPenalty float64  `json:"repeat_penalty,omitempty"`
	NumThread     int      `json:"num_thread,omitempty"`
	Stop          []string `json:"stop,omitempty"`
}

type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

func GetDeepSeekConfig() Options {
	ops := Options{
		Temperature:   0.7,  // Более низкая температура для детерминированных ответов (Стандартное значение)
		TopP:          0.9,  // Используем top-p sampling (Стандартное значение)
		TopK:          40,   // Ограничиваем выбор топ-40 токенов (Стандартное значение)
		NumPredict:    6144, // Ограничиваем длину ответа (Достаточно для 15 вопросов)
		RepeatPenalty: 1.1,  // Стандартное значение
		NumThread:     14,
	}
	numThread := config.Conf.AI.Ollama.NumThread
	if numThread > 0 {
		ops.NumThread = numThread
	}

	repeatPenalty := config.Conf.AI.Ollama.RepeatPenalty
	if repeatPenalty > 0 {
		ops.RepeatPenalty = repeatPenalty
	}

	numPredict := config.Conf.AI.Ollama.NumPredict
	if numPredict > 0 {
		ops.NumPredict = numPredict
	}

	topK := config.Conf.AI.Ollama.TopK
	if topK > 0 {
		ops.TopK = topK
	}

	topP := config.Conf.AI.Ollama.TopP
	if topP > 0 {
		ops.TopP = topP
	}

	temperature := config.Conf.AI.Ollama.Temperature
	if temperature > 0 {
		ops.Temperature = temperature
	}
	return ops
}
