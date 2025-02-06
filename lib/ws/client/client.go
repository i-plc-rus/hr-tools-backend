package wsclient

import (
	"fmt"
	// connectionhub "hr-tools-backend/lib/ws/hub/connection-hub"
	// wsmodels "hr-tools-backend/models/ws"
	// "time"

	"github.com/gofiber/contrib/websocket"
	log "github.com/sirupsen/logrus"
)

func NewClient(userID string, c *websocket.Conn) *WsClient {
	return &WsClient{
		conn:   c,
		userID: userID,
	}
}

type WsClient struct {
	conn   *websocket.Conn
	userID string
}

type Message struct {
	MessageType int
	Msg         []byte
}

var closeCodes []int

func init() {
	for i := websocket.CloseNormalClosure; i <= websocket.CloseTLSHandshake; i++ {
		closeCodes = append(closeCodes, i)
	}
}

func (c *WsClient) Dispatch() {
	for {
		if c.conn == nil {
			return
		}
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, closeCodes...) {
				log.WithError(err).Error("ошибка получения сообщения")
			}
			break
		}
		log.WithField("ws_message", fmt.Sprintf("%+v", data)).Debug("ws-msg")
	}
}
