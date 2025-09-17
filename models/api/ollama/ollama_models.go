package ollamamodels

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
	Stop          []string `json:"stop,omitempty"`
}

type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}


func GetDeepSeekConfig() Options {
	return Options{
		Temperature:   0.7,  // Более низкая температура для детерминированных ответов (Стандартное значение)
		TopP:          0.9,  // Используем top-p sampling (Стандартное значение)
		TopK:          40,   // Ограничиваем выбор топ-40 токенов (Стандартное значение)
		NumPredict:    2000, // Ограничиваем длину ответа (Достаточно для 15 вопросов)
		RepeatPenalty: 1.1,  // Стандартное значение
		Stop:          []string{"повторяю", "как уже говорил", "еще раз"},
	}
}
