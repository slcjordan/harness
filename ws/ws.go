package ws

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type Listener[T any] interface {
	Message(string, string, T)
	Closed(string)
}

type Server[T any] struct {
	OriginPatterns []string
	Conns          map[string]*websocket.Conn
	Listener       Listener[T]
}

func (s *Server[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Accept the WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: s.OriginPatterns,
	})
	if err != nil {
		log.Println("accept error:", err)
		return
	}
	id := uuid.New().String()
	defer s.Listener.Closed(id)
	defer conn.Close(websocket.StatusNormalClosure, "done")

	log.Println("Client connected")

	ctx := r.Context()

	// Set up ping ticker
	pingInterval := time.Second * 10
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()

	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Ping the client to keep the connection alive
				err := conn.Ping(ctx)
				if err != nil {
					log.Println("ping error:", err)
					conn.Close(websocket.StatusGoingAway, "ping failed")
					return
				}
			case <-pingCtx.Done():
				return
			}
		}
	}()

	// Read messages and echo them back
	for {
		var msg map[string]T
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			log.Println("read error:", err)
			break
		}
		for topic, input := range msg {
			s.Listener.Message(id, topic, input)
		}

	}
}

func (s *Server[T]) MaybeSend(id string, topic string, output T) {
	conn, ok := s.Conns[id]
	if !ok {
		// log.Errorf(conn.Context(), "could not find")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := wsjson.Write(ctx, conn, output)
	if err != nil {
		// log.Println("write error:", err)
	}
}
