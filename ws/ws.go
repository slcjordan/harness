package ws

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/logger"
)

type Server struct {
	OriginPatterns []string
	mu             sync.Mutex
	Conns          map[string]*websocket.Conn
	Model          harness.Daemon
	Command        harness.Handler[[]byte, struct{}]
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
	s.setConn(id, conn)
	defer s.deleteConn(id)
	s.Model.Attach(id)
	defer s.Model.Unattach(id)
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
		go func() {
			_, err := s.Command.Handle(ctx, input)
			if err != nil {
				log.Println("read error:", err)
			}
		}()
	}
}

func (s *Server) getConn(id string) (*websocket.Conn, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, ok := s.Conns[id]
	return result, ok
}

func (s *Server) deleteConn(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Conns, id)
}

func (s *Server) setConn(id string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Conns[id] = conn
}

func (s *Server) Notify(id string, output []byte) {
	conn, ok := s.getConn(id)
	if !ok {
		logger.Errorf(context.TODO(), "could not find ws conn: %q", id)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := conn.Write(ctx, websocket.MessageText, output)
	if err != nil {
		logger.Infof(context.TODO(), "could not send to ws %q: %s", id, err)
		return
	}
}
