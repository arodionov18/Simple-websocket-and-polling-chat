package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

var chatHistory []Message
var chatChannel = make(chan Message)


var m *melody.Melody



func PollingServe(context *gin.Context) {
	log.Println("Got polling request")

	var json Request
	if err := context.ShouldBindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if json.Type == MSG_SEND {
		str := PrepareSendMessage(*json.Name, *json.Text)
		chatHistory = append(chatHistory, Message{Name: *json.Name, Text: *json.Text})
		m.Broadcast(str)
		chatChannel <- Message{Name: *json.Name, Text: *json.Text}

		reply := Reply{Ok: true, Messages: []Message{}}
		context.JSON(http.StatusOK, reply)
		return
	} else if json.Type == MSG_GET {
		reply := Reply{Ok: true, Messages: chatHistory}
		context.JSON(http.StatusOK, reply)
		chatHistory = nil
	}
}

func LongPollingServe(context *gin.Context) {
	log.Println("Got long polling request")
	var json Request
	if err := context.ShouldBindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	 if json.Type == MSG_SEND {
		str := PrepareSendMessage(*json.Name, *json.Text)
		chatHistory = append(chatHistory, Message{Name: *json.Name, Text: *json.Text})
		chatChannel <- Message{Name: *json.Name, Text: *json.Text}
		m.Broadcast(str)

		reply := Reply{Ok: true, Messages: []Message{}}
		context.JSON(http.StatusOK, reply)
		return
	} else if json.Type == MSG_GET {
		select {
		case msg := <- chatChannel:
			reply := Reply{Ok: true, Messages: []Message{msg}}
			context.JSON(http.StatusOK, reply)
		}
	}
}

func PrepareSendError(errorText string) []byte {
	reply := Reply{
		Ok: false,
		Error: &errorText,
	}
	marshaled, err := json.Marshal(reply)
	if err != nil {
		log.Println("marshal:", err)
	}

	return marshaled
}

func PrepareSendMessage(name string, text string) []byte {
	reply := Reply{
		Ok:       true,
		Error:    nil,
		Messages: []Message{{name, text}},
	}

	marshaled, err := json.Marshal(reply)
	if err != nil {
		log.Println("marshal:", err)
	}

	return marshaled
}

func ServeWebsocket(s *melody.Session, data []byte) {
	log.Println("Got websocket request")

	var msg Request
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Println("failed to unmarshal msg", err)
		return
	}

	if msg.Type == MSG_SEND {
		log.Println("Send websocket request")

		str := PrepareSendMessage(*msg.Name, *msg.Text)
		chatHistory = append(chatHistory, Message{Name: *msg.Name, Text: *msg.Text})
		chatChannel <- Message{Name: *msg.Name, Text: *msg.Text}
		m.Broadcast(str)
	}
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
