package main

import (
	"context"
	"net/http"
	"os"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/exec"
	"github.com/slcjordan/harness/json"
	"github.com/slcjordan/harness/logger"
	"github.com/slcjordan/harness/ws"
)

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// data, err := os.ReadFile("/tmp/portal_test.key")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(data))
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fs := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(config.HTTPServer.FileRoot)).ServeHTTP(w, r)
	})
	enc := &json.InteractiveCommandEncoder{}
	daemon := exec.StartInteractive(ctx, enc, "cat")
	enc.Command = daemon
	ws := &ws.Server{
		Conns:   make(map[string]*websocket.Conn),
		Model:   daemon,
		Command: enc,
	}
	enc.Listener = ws
	r.Handle("/app/*", http.StripPrefix("/app/", fs))
	r.Handle("/app/ws/*", http.StripPrefix("/app/ws/", ws))
	/*
		r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/app/index.html", http.StatusMovedPermanently)
		})
	*/

	c := cli.NewCommand("serve", "run http server", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
		logger.Infof(ctx, "listening at %q\n", config.HTTPServer.Addr)
		return http.ListenAndServe(config.HTTPServer.Addr, r)
	}), cli.WithHTTPServerFlags)

	err := c.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
