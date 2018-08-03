package user_management

import (
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/phanletrunghieu/chat-app-golang/domain"
)

var instance *UserManagement

type UserManagement struct {
	listClients map[*domain.User]*websocket.Conn
}

func GetInstace() *UserManagement {
	if instance == nil {
		instance = &UserManagement{
			listClients: make(map[*domain.User]*websocket.Conn),
		}
	}
	return instance
}

func (um *UserManagement) GetListUser() []*domain.User {
	listUser := make([]*domain.User, 0)
	for client, _ := range um.listClients {
		listUser = append(listUser, client)
	}

	return listUser
}

func (um *UserManagement) SendBroadcast(message interface{}) {
	for _, ws := range um.listClients {
		ws.WriteJSON(message)
	}
}

func (um *UserManagement) Connect(user *domain.User, ws *websocket.Conn) {
	um.listClients[user] = ws
}

func (um *UserManagement) Disconnect(user *domain.User) {
	log.Println(user.Name, "disconected")

	l1 := len(um.listClients)
	delete(um.listClients, user)
	l2 := len(um.listClients)

	if l1 == l2 {
		for u := range um.listClients {
			delete(um.listClients, u)
		}
	}

	um.SendBroadcast(&domain.Message{
		Type: 4,
		Data: user,
	})
}

func (um *UserManagement) GetSocketByUId(user *domain.User) (*websocket.Conn, error) {
	for client, ws := range um.listClients {
		if client.ID == user.ID {
			return ws, nil
		}
	}

	return nil, errors.New("User not found")
}
