package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	"websocket_chat/message"
)

func runClient() {
	addr := "localhost:8088"

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, byteMsg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var msg message.Reply
			err = json.Unmarshal(byteMsg, &msg)
			if err != nil {
				log.Println("unmarshal:", err)
				return
			}
			if !msg.Ok {
				log.Println("reply:", msg.Error)
				continue
			}
			for _, ms := range msg.Messages {
				fmt.Printf("%s: %s\n", ms.Name, ms.Text)
			}

		}
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			splites := strings.FieldsFunc(scanner.Text(), func(r rune) bool {
				return r == '|'
			})
			if len(splites) < 2 {
				continue
			}
			if splites[0] == "Reg" {
				// Send Reg request
				req := message.NewRegisterRequest(splites[1])
				marshaled, err := json.Marshal(req)
				if err != nil {
					log.Println("marshal:", err)
					return
				}
				err = c.WriteMessage(websocket.TextMessage, marshaled)
				if err != nil {
					log.Println("send:", err)
					return
				}
			} else if splites[0] == "Send" {
				// Send message
				req := message.NewSendRequest(splites[1], splites[2])
				marshaled, err := json.Marshal(req)
				if err != nil {
					log.Println("marshal:", err)
					return
				}
				err = c.WriteMessage(websocket.TextMessage, marshaled)
				if err != nil {
					log.Println("send:", err)
					return
				}
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			}
			return
		}
	}

}