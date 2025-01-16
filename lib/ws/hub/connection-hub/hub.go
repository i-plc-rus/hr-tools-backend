package connectionhub

import (
	"hr-tools-backend/db"
	pushdatastore "hr-tools-backend/lib/space/push/data-store"
	wsmodels "hr-tools-backend/models/ws"

	"github.com/gofiber/contrib/websocket"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	AddClient(userID string, conn *websocket.Conn)
	DeleteClient(userID string)
	SendMessage(msg wsmodels.ServerMessage)
	SendClose(userID string)
	IsConnected(userID string) bool
}

var Instance Provider

func Init() {
	Instance = &impl{
		clients: map[string]clientSession{},
		store:   pushdatastore.NewInstance(db.DB),
	}
}

type impl struct {
	clients map[string]clientSession //map[userID]
	store   pushdatastore.Provider
}

func (i *impl) DeleteClient(userID string) {
	sess, ok := i.clients[userID]
	if !ok {
		return
	}
	delete(i.clients, userID)
	sess.stop()
	close(sess.sendCh)
}

func (i *impl) AddClient(userID string, conn *websocket.Conn) {
	oldSess, ok := i.clients[userID]
	if ok {
		oldSess.stop()
	}
	i.clients[userID] = newSession(conn)
	go i.sendDelayedMessages(userID)
}

func (i *impl) SendMessage(msg wsmodels.ServerMessage) {
	userID := msg.ToUserID
	sess, ok := i.clients[userID]
	if ok {
		sess.sendCh <- msg
	}
}

func (i *impl) SendClose(userID string) {
	sess, ok := i.clients[userID]
	if ok {
		sess.stop()
	}
}

func (i *impl) IsConnected(userID string) bool {
	sess, ok := i.clients[userID]
	if !ok || sess.conn == nil || sess.conn.Conn == nil {
		return false
	}
	return true
}

func (i *impl) sendDelayedMessages(userID string) {
	logger := log.WithField("user_id", userID)
	list, err := i.store.List(userID)
	if err != nil {
		logger.WithError(err).Error("ошибка получения списка не отправленных событий")
		return
	}
	sendedIDs := []string{}
	for _, item := range list {
		if i.IsConnected(userID) {
			msg := wsmodels.ServerMessage{
				ToUserID: userID,
				Time:     item.CreatedAt.Format("02.01.2006 15:04:05"),
				Code:     string(item.Code),
				Msg:      item.Msg,
			}
			i.SendMessage(msg)
			sendedIDs = append(sendedIDs, item.ID)
		}
	}
	if len(sendedIDs) > 0 {
		err = i.store.Delete(sendedIDs)
		if err != nil {
			logger.WithError(err).Error("ошибка удаления отправленных событий")
			return
		}
	}
}
