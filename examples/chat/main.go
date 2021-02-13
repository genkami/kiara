package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/genkami/kiara"
	adapter "github.com/genkami/kiara/adapter/redis"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

const (
	messageChSize  = 10
	maxMessageSize = 512

	topicLobby = "room:lobby"
)

var (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	publishTimeout = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	From string `json:"from"`
	Body string `json:"body"`
}

type Server struct {
	pubsub *kiara.PubSub
}

type Session struct {
	conn     *websocket.Conn
	pubsub   *kiara.PubSub
	sub      *kiara.Subscription
	messages <-chan *Message
	done     chan struct{}
}

func main() {
	var redisAddr string
	var bindAddr string
	flag.StringVar(&redisAddr, "redis-addr", "localhost:6379", "Redis address")
	flag.StringVar(&bindAddr, "bind-addr", ":8080", "bind address")
	flag.Parse()

	var err error
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	pubsub := kiara.NewPubSub(adapter.NewAdapter(redisClient))
	defer pubsub.Close()

	server := &Server{pubsub: pubsub}

	http.HandleFunc("/", server.serveHome)
	http.HandleFunc("/ws", server.serveWs)
	err = http.ListenAndServe(bindAddr, nil)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (_ *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	http.ServeFile(w, r, "index.html")
}

func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	messages := make(chan *Message, messageChSize)
	sub, err := s.pubsub.Subscribe(topicLobby, messages)
	if err != nil {
		log.Println(err)
		conn.WriteMessage(websocket.CloseMessage, []byte{})
		return
	}

	sess := &Session{
		conn:     conn,
		pubsub:   s.pubsub,
		sub:      sub,
		messages: messages,
		done:     make(chan struct{}),
	}
	go sess.handleIncoming()
	go sess.handleOutgoing()
}

func (s *Session) handleIncoming() {
	defer func() {
		err := s.sub.Unsubscribe()
		if err != nil {
			log.Println(err)
		}
		s.conn.Close()
	}()
	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msg := &Message{}
		err := s.conn.ReadJSON(msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			return
		}
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
			defer cancel()
			err = s.pubsub.Publish(ctx, topicLobby, msg)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func (s *Session) handleOutgoing() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := s.sub.Unsubscribe()
		if err != nil {
			log.Println(err)
		}
		s.conn.Close()
	}()

	for {
		select {
		case msg := <-s.messages:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := s.conn.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
