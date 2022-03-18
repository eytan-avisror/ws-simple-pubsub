package pubsub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	OpPublish     string = "publish"
	OpSubscribe   string = "subscribe"
	OpUnsubscribe string = "unsubscribe"
	OpRemove      string = "remove"
	OpList        string = "list"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var server = NewPubSubServer()

func WSHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	// generate unique ID for client
	clientID := uuid.Must(uuid.NewRandom()).String()
	message := fmt.Sprintf("server: new client %v", clientID)
	server.SendMessage(conn, message, 1)

	// message handling
	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			server.RemoveClient(clientID)
			log.Errorf("failed to read message from socket: %v", err)
			return
		}
		server.ProcessMessage(clientID, conn, msgType, payload)
	}
}

type Server interface {
	PublishMessage(topic string, message []byte, msgType int)
	SendMessage(c *websocket.Conn, message string, msgType int)
	RemoveClient(clientID string)
	Unsubscribe(clientID, topic string)
	ListTopics(conn *websocket.Conn)
	Subscribe(clientID, topic string, conn *websocket.Conn)
	ProcessMessage(clientID string, conn *websocket.Conn, messageType int, payload []byte)
}

func NewPubSubServer() Server {
	return &PubSubServer{
		RWMutex:       sync.RWMutex{},
		Subscriptions: NewSubscriptionList(),
	}
}

type PubSubServer struct {
	sync.RWMutex
	Subscriptions SubscriptionList
}

// SubscriptionList maps Topics to a ClientList
type SubscriptionList struct {
	Items map[string]ClientList
}

func NewSubscriptionList() SubscriptionList {
	return SubscriptionList{
		Items: make(map[string]ClientList),
	}
}

// ClientList maps a ClientID to a Websocket connection
type ClientList struct {
	Items map[string]*websocket.Conn
}

type Message struct {
	Operation string `json:"op"`
	Topic     string `json:"topic"`
	Message   string `json:"message"`
}

func (s *PubSubServer) ProcessMessage(clientID string, conn *websocket.Conn, msgType int, payload []byte) {
	m := Message{}
	if err := json.Unmarshal(payload, &m); err != nil {
		s.SendMessage(conn, "server: failed to unmarshal payload", 1)
	}

	switch strings.ToLower(m.Operation) {
	case OpPublish:
		s.PublishMessage(m.Topic, []byte(m.Message), msgType)
	case OpSubscribe:
		s.Subscribe(clientID, m.Topic, conn)
	case OpUnsubscribe:
		s.Unsubscribe(clientID, m.Topic)
	case OpRemove:
		s.RemoveClient(clientID)
	case OpList:
		s.ListTopics(conn)
	default:
		err := fmt.Sprintf("server: unknown operation '%v'", m.Operation)
		s.SendMessage(conn, err, 1)
	}
}

func (s *PubSubServer) GetSubscriptions() SubscriptionList {
	return s.Subscriptions
}

func (s *PubSubServer) SendMessage(c *websocket.Conn, message string, msgType int) {
	c.WriteMessage(msgType, []byte(message))
}

// ListTopics returns all available topics
func (s *PubSubServer) ListTopics(conn *websocket.Conn) {
	var (
		subscriptions = s.GetSubscriptions()
	)
	log.Infof("listing all topics")

	s.RLock()
	defer s.RUnlock()

	if len(subscriptions.Items) < 1 {
		s.SendMessage(conn, "server has no topics, create one!", 1)
		return
	}

	for topic, subscribers := range subscriptions.Items {
		msg := fmt.Sprintf("topic %v has %v subscribers", topic, len(subscribers.Items))
		s.SendMessage(conn, msg, 1)
	}
}

// Unsubscribe a client from all topics
func (s *PubSubServer) RemoveClient(clientID string) {
	var (
		subscriptions = s.GetSubscriptions()
	)
	log.Infof("removing subscriber '%v' from all topics", clientID)

	s.Lock()
	defer s.Unlock()

	for topic, subscribers := range subscriptions.Items {

		if _, ok := subscribers.Items[clientID]; ok {
			// found subscription for client
			delete(subscribers.Items, clientID)
			log.Infof("removed subscriptions for client '%v' from topic '%v'", clientID, topic)
		}
	}
}

// Unsubscribe a client from a specific topic
func (s *PubSubServer) Unsubscribe(clientID, topic string) {
	var (
		subscriptions = s.GetSubscriptions()
	)
	log.Infof("removing subscriber '%v' from topic '%v'", clientID, topic)

	s.Lock()
	defer s.Unlock()

	if subscribers, ok := subscriptions.Items[topic]; ok {
		// found subscription for client
		delete(subscribers.Items, clientID)
		log.Infof("removed subscriptions for client '%v' from topic '%v'", clientID, topic)
	}
}

// Publish a message to all subscribers
func (s *PubSubServer) PublishMessage(topic string, message []byte, msgType int) {
	var (
		subscriptions = s.GetSubscriptions()
	)
	log.Infof("publishing message in topic '%v'", topic)

	s.RLock()
	defer s.RUnlock()

	if subscribers, ok := subscriptions.Items[topic]; ok {
		for _, conn := range subscribers.Items {
			s.SendMessage(conn, string(message), msgType)
		}
	} else {
		// otherwise create the new topic but don't subscribe to it
		subscriptions.Items[topic] = ClientList{}
	}
}

func (s *PubSubServer) Subscribe(clientID, topic string, conn *websocket.Conn) {
	var (
		subscriptions = s.GetSubscriptions()
	)
	log.Infof("adding subscriber '%v' to topic '%v'", clientID, topic)

	s.Lock()
	defer s.Unlock()

	// Add subscriber to topic if it exists
	if subscribers, ok := subscriptions.Items[topic]; ok {
		subscribers.Items[clientID] = conn
	} else {
		// otherwise create the new topic with the subscriber
		subscriptions.Items[topic] = ClientList{
			Items: map[string]*websocket.Conn{
				clientID: conn,
			},
		}
	}
}
