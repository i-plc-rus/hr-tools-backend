package connectionhub

import (
	"context"
	"github.com/gofiber/contrib/websocket"
	log "github.com/sirupsen/logrus"
	"time"
)

type clientSession struct {
	conn *websocket.Conn

	// Outbound mesages, buffered.
	// The content must be serialized in format suitable for the session.
	sendCh chan any
	stop   func()
}

func newSession(conn *websocket.Conn) clientSession {
	ctx, cancelFn := context.WithCancel(context.TODO())
	sess := clientSession{
		stop:   cancelFn,
		conn:   conn,
		sendCh: make(chan any, 1), // buffered,
	}
	go sess.startSend(ctx)
	return sess
}

func (s clientSession) startSend(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.close()
			return
		case msg, opened := <-s.sendCh:
			if !opened {
				return
			}
			_, err := s.send(s.conn, msg)
			if err != nil {
				log.WithError(err).Error("ошибка отправки сообщения")
			}
		}
	}
}

func (s clientSession) send(conn *websocket.Conn, msg interface{}) (bool, error) {
	if conn.Conn == nil {
		return false, nil
	}
	err := conn.WriteJSON(msg)
	if err != nil {
		return false, err
	}
	log.Infof("отправлено сообщение: %s", msg)
	return true, nil
}

func (s clientSession) close() {
	if s.conn == nil || s.conn.Conn == nil {
		return
	}
	err := s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Millisecond))
	if err != nil {
		log.WithError(err).Error("cant close")
	}
}
