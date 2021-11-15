package main

import (
	"errors"
	"log"
	"net/http"

	"websocket_chat/message"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"encoding/json"
)

var clients = map[string]struct{}{}
var chatHistory []message.Message


var m *melody.Melody


func RegisterUser(name string) error {
	_, ok := clients[name]
	if ok {
		return errors.New("choose different Name")
	}
	clients[name] = struct{}{}
	return nil
}

func PollingServe(context *gin.Context) {
	log.Println("Got polling request")

	var json message.Request
	if err := context.ShouldBindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if json.Type == message.MSG_REGISTER {
		err := RegisterUser(*json.Name)
		if err != nil {
			text := err.Error()
			reply := message.Reply{Ok: false, Error: &text}
			context.JSON(http.StatusBadRequest, reply)
		}
		reply := message.Reply{Ok: true, Messages: []message.Message{{Name: "", Text: "You are registered"}}}
		context.JSON(http.StatusOK, reply)
		return
	} else if json.Type == message.MSG_SEND {
		_, ok := clients[*json.Name]
		if !ok {
			text := "you must be registered before sending message"
			reply := message.Reply{Ok: false, Error: &text}
			context.JSON(http.StatusBadRequest, reply)
		}
		str := SendMessage(*json.Name, *json.Text)
		chatHistory = append(chatHistory, message.Message{Name: *json.Name, Text: *json.Text})
		m.Broadcast(str)
		reply := message.Reply{Ok: true, Messages: []message.Message{{Name: *json.Name, Text: *json.Text}}}
		context.JSON(http.StatusOK, reply)
		return
	} else if json.Type == message.MSG_GET {
		reply := message.Reply{Ok: true, Messages: chatHistory}
		context.JSON(http.StatusOK, reply)
		chatHistory = nil
	}
}

func LongPollingServe(context *gin.Context) {
	context.Writer.WriteString("Long Hello World!")
}

func SendError(errorText string) []byte {
	reply := message.Reply{
		Ok: false,
		Error: &errorText,
	}
	marshaled, err := json.Marshal(reply)
	if err != nil {
		log.Println("marshal:", err)
	}
	return marshaled
}

func SendMessage(name string, text string) []byte {
	reply := message.Reply{
		Ok:       true,
		Error:    nil,
		Messages: []message.Message{{name, text}},
	}
	marshaled, err := json.Marshal(reply)
	if err != nil {
		log.Println("marshal:", err)
	}
	return marshaled
}

func ServeWebsocket(s *melody.Session, data []byte) {
	log.Println("Got websocket request")
	var msg message.Request
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Println("failed to unmarshal msg", err)
		return
	}
	if msg.Type == message.MSG_REGISTER {
		log.Println("Register websocket request")
		err := RegisterUser(*msg.Name)
		if err != nil {
			s.Write(SendError(err.Error()))
			return
		}
		s.Write(SendMessage("", "You are registered"))
	} else if msg.Type == message.MSG_SEND {
		log.Println("Send websocket request")
		_, ok := clients[*msg.Name]
		if !ok {
			s.Write(SendError("You must register before sending messages"))
			return
		}

		str := SendMessage(*msg.Name, *msg.Text)
		chatHistory = append(chatHistory, message.Message{Name: *msg.Name, Text: *msg.Text})
		m.Broadcast(str)
	}

	//s.Write([]byte("Websocket HelloWorld!"))
}


func runServer() {
	r := gin.Default()
	m = melody.New()

	r.Any("/", PollingServe)

	r.Any("/longPolling", LongPollingServe)

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(ServeWebsocket)

	r.Run(*addr)
}
