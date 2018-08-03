package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/phanletrunghieu/chat-app-golang/domain"
	"github.com/phanletrunghieu/chat-app-golang/domain/user_management"
	uuid "github.com/satori/go.uuid"
)

func main() {
	user_management := user_management.GetInstace()
	var channelMessage = make(chan domain.Message)

	// Create a simple file server
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.Handle("/ws", MiddlewareWS(HandleWS(user_management, channelMessage)))

	go HandleMessages(user_management, channelMessage)

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func MiddlewareWS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing middleware")

		query := r.URL.Query()
		tokens, ok := query["token"]
		names, ok := query["name"]
		if !ok ||
			len(tokens) < 1 ||
			len(names) < 1 {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		user := &domain.User{
			ID:    uuid.NewV4().String(),
			Token: tokens[0],
			Name:  names[0],
		}

		userJson, _ := json.Marshal(user)
		r.Header.Add("x-user-info", string(userJson))

		if user.Token != "xxx" {
			http.Error(w, "token invalid", 400)
		}
		next.ServeHTTP(w, r)
	})
}

func HandleWS(user_management *user_management.UserManagement, channelMessage chan domain.Message) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{}
		// Upgrade initial GET request to a websocket
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		// Make sure we close the connection when the function returns
		defer ws.Close()

		// Register our new client
		ruser := r.Header.Get("x-user-info")
		user := &domain.User{}
		json.Unmarshal([]byte(ruser), user)

		user_management.SendBroadcast(&domain.Message{
			Data: user,
			Type: 3,
		})

		user_management.Connect(user, ws)

		log.Println(user.Name, "connected")

		ws.WriteJSON(domain.Message{
			From: &domain.User{ID: "-1"},
			Data: user.ID,
			Type: 0,
		})

		for {
			var msg domain.Message
			// Read in a new message as JSON and map it to a Message object
			err := ws.ReadJSON(&msg)
			if err != nil {
				user_management.Disconnect(user)
				break
			}

			msg.From = user

			log.Println(user.Name, "send:", msg)

			// Send the newly received message to the message channel
			switch msg.Type {
			case 1:
				channelMessage <- msg
			case 2: //get list user
				listUser := user_management.GetListUser()
				mg := &domain.Message{
					Type: 2,
					Data: listUser,
				}
				ws.WriteJSON(mg)
			}
		}
	})
}

func HandleMessages(user_management *user_management.UserManagement, channelMessage chan domain.Message) {
	for {
		// Grab the next message from the message channel
		msg := <-channelMessage
		ws, err := user_management.GetSocketByUId(msg.To)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		err = ws.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			ws.Close()
			user_management.Disconnect(msg.To)
		}
	}
}
