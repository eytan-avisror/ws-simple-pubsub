package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/eytan-avisror/ws-simple-pubsub/pkg/pubsub"

	log "github.com/sirupsen/logrus"
)

var (
	serverEndpoint string
	serverPort     string
)

func main() {

	flag.StringVar(&serverEndpoint, "bind-addr", "localhost", "the address to bind the server to")
	flag.StringVar(&serverPort, "bind-port", "8080", "the port to bind the server to")
	flag.Parse()

	ep := fmt.Sprintf("%v:%v", serverEndpoint, serverPort)

	log.Info("registering websocket handler '/ws'")
	http.HandleFunc("/ws", pubsub.WSHandler)

	log.Infof("starting server on %v", ep)
	err := http.ListenAndServe(ep, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
