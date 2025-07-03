package ws

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/slcjordan/harness"
)

type Server struct {
	OriginPatterns []string
	Conns          map[string]*websocket.Conn
	Listener       harness.Listener[[]byte]
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Accept the WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: s.OriginPatterns,
	})
	if err != nil {
		log.Println("accept error:", err)
		return
	}
	id := uuid.New().String()
	defer s.Listener.Done(id)
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
		_, input, err := conn.Read(ctx)
		if err != nil {
			log.Println("read error:", err)
			break
		}
		go s.Listener.Receive(id, input)
	}
}

func (s *Server) MaybeSend(id string, output []byte) {
	conn, ok := s.Conns[id]
	if !ok {
		// log.Errorf(conn.Context(), "could not find")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := conn.Write(ctx, websocket.MessageText, output)
	if err != nil {
		// log.Println("write error:", err)
	}
}
