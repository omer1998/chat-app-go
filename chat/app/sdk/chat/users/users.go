package users

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"
	"github.com/omer1998/chat-app-go.git/chat/foundation/logger"
)

type Users struct {
	users  map[uuid.UUID]chat.User
	muUser sync.RWMutex
	log    *logger.Logger
}

func NewUsers(log *logger.Logger) *Users {
	return &Users{
		log:   log,
		users: make(map[uuid.UUID]chat.User),
	}
}

func (u *Users) AddUser(usr chat.User) error {
	u.muUser.Lock()
	defer u.muUser.Unlock()
	_, exists := u.users[usr.Id]
	if exists {
		return fmt.Errorf("user already exists")
	}

	u.log.Debug(context.Background(), "user added", "status", "success", "id", usr.Id, "name", usr.Name)
	u.users[usr.Id] = usr
	return nil
}

func (u *Users) RemoveUser(cxt context.Context, id uuid.UUID) error {
	u.muUser.Lock()
	defer u.muUser.Unlock()
	v, exist := u.users[id]
	if !exist {
		u.log.Info(cxt, "remove user", "id", id, "status", "does not exist")
		return chat.ErrUserNotExist
	}
	v.Conn.Close()
	delete(u.users, id)
	u.log.Info(cxt, "remove user", "status", "success", "id", id, "name", v.Name)
	return nil
}

func (u *Users) RetrieveUser(id uuid.UUID) (chat.User, error) {
	u.muUser.RLock()
	defer u.muUser.RUnlock()
	usr, exist := u.users[id]
	if !exist {
		return chat.User{}, chat.ErrUserNotExist
	}
	return usr, nil
}

func (u *Users) Connections() map[uuid.UUID]chat.Connection {
	u.muUser.RLock()
	defer u.muUser.RUnlock()
	connections := make(map[uuid.UUID]chat.Connection)
	for k, v := range u.users {
		connections[k] = chat.Connection{
			Conn:     v.Conn,
			LastPing: v.LastPing,
			LastPong: v.LastPong,
		}
	}
	return connections
}

func (u *Users) UpdateLastPingTime(usrId uuid.UUID) error {
	u.muUser.Lock()
	defer u.muUser.Unlock()
	usr, exist := u.users[usrId]
	if !exist {
		return chat.ErrUserNotExist
	}
	usr.LastPing = time.Now()
	u.users[usr.Id] = usr
	u.log.Debug(context.Background(), "update last ping time", "status", "success", "lastPingTime", usr.LastPing.String())

	return nil
}
func (u *Users) UpdateLastPongTime(usrId uuid.UUID) error {
	u.muUser.Lock()
	defer u.muUser.Unlock()
	usr, exist := u.users[usrId]
	if !exist {
		return chat.ErrUserNotExist
	}
	usr.LastPong = time.Now()
	u.users[usr.Id] = usr
	u.log.Debug(context.Background(), "update last pong time", "status", "success", "lastPongTime", usr.LastPong.String())

	return nil
}
