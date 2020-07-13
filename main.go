package main

import (
	"log"
	"stream-service/server"
)

func main() {
	log.SetFlags(log.Ldate | log.Lshortfile)
	// server.StartWebRTC()
	// server.StartHTTP()
	server.StartWebSocket()
}
