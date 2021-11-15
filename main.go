package main

import "flag"

var addr = flag.String("addr", ":8088", "http service address")
var typ = flag.String("type", "server", "")

func main() {

	flag.Parse()
	if *typ == "server" {
		runServer()
	} else if *typ == "client" {
		runClient()
	} else if *typ == "polling" {
		runPollingClient("/")
	} else if *typ == "longpolling" {
		runPollingClient("/longPolling")
	}
}
