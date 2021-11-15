package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

func sendRequestAndPrintReply(request []byte, u url.URL,) error {
	responseBody := bytes.NewBuffer(request)
	resp, err := http.Post(u.String(), "application/json", responseBody)
	if err != nil {
		log.Println("send:", err)
		return err
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil:", err)
	}

	var reply Reply
	err = json.Unmarshal(res, &reply)
	if err != nil {
		log.Println("unmarshal:", err)
	}

	if !reply.Ok {
		fmt.Println(reply.Error)
	} else {
		for _, msg := range reply.Messages {
			fmt.Printf("%s: %s\n", msg.Name, msg.Text)
		}
	}
	resp.Body.Close()
	return nil
}

func runPollingClient(path string) {
	addr := "localhost:8088"
	pollTimeout := time.Second * 5

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "http", Host: addr, Path: path}
	log.Printf("connecting to %s", u.String())

	ticker := time.NewTicker(pollTimeout)
	defer ticker.Stop()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			splits := strings.FieldsFunc(scanner.Text(), func(r rune) bool {
				return r == '|'
			})
			if len(splits) < 2 {
				continue
			}

			if splits[0] == "Send" {
				req := NewSendRequest(splits[1], splits[2])
				marshaled, err := json.Marshal(req)
				if err != nil {
					log.Println("marshal:", err)
					return
				}

				if err := sendRequestAndPrintReply(marshaled, u); err != nil {
					return
				}
			}
		}
	}()

	for {
		select {
		case _ = <-ticker.C:
			body := NewGetRequest()
			postBody, err := json.Marshal(body)
			if err != nil {
				log.Println("marshall:", err)
				return
			}

			responseBody := bytes.NewBuffer(postBody)
			resp, err := http.Post(u.String(), "application/json", responseBody)
			if err != nil {
				log.Println("poll:", err)
				return
			}

			defer resp.Body.Close()

			res, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("ioutil:", err)
			}

			var reply Reply
			err = json.Unmarshal(res, &reply)
			if err != nil {
				log.Println("unmarshal:", err)
			}

			if !reply.Ok {
				fmt.Println(reply.Error)
			} else {
				for _, msg := range reply.Messages {
					fmt.Printf("%s: %s\n", msg.Name, msg.Text)
				}
			}

		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
