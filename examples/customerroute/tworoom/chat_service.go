package tworoom

import (
	"github.com/acoderup/core/logger"
	"github.com/acoderup/nano"
	"github.com/acoderup/nano/component"
	"github.com/acoderup/nano/examples/cluster/protocol"
	"github.com/acoderup/nano/session"
)

type ChatRoomService struct {
	component.Base
	group *nano.Group
}

func newChatRoomService() *ChatRoomService {
	return &ChatRoomService{
		group: nano.NewGroup("all-users"),
	}
}

func (rs *ChatRoomService) JoinRoom(s *session.Session, msg *protocol.JoinRoomRequest) error {
	return rs.group.Add(s)
}

type SyncMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (rs *ChatRoomService) SyncMessage(s *session.Session, msg *SyncMessage) error {
	// Sync message to all members in this room
	return rs.group.Broadcast("onMessage", msg)
}

func (rs *ChatRoomService) userDisconnected(s *session.Session) {
	if err := rs.group.Leave(s); err != nil {
		logger.Logger.Tracef("Remove user from group failed uid[%v] err[%v]", s.UID(), err)
		return
	}
	logger.Logger.Tracef("User session disconnected uid[%v]", s.UID())
}
