package avitoapimodels

type NewMessageRequest struct {
	Message MessageData `json:"message"`
	Type    string      `json:"type"`
}

type MessageData struct {
	Text string `json:"text"`
}

type MessageResponse struct {
	Messages []MessageItem `json:"messages"`
}

type MessageItem struct {
	ID        string         `json:"id"`
	Created   int64          `json:"created"`
	Type      string         `json:"type"`
	Direction string         `json:"direction"`
	IsRead    bool           `json:"is_read"`
	AuthorID  int            `json:"author_id"`
	Content   MessageContent `json:"content"`
}

type MessageContent struct {
	Text string `json:"text"`
}

type ChatInfo struct {
	ID          string      `json:"id"`
	Created     int64       `json:"created"`
	LastMessage MessageItem `json:"last_message"`
}
