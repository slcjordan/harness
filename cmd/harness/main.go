package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/ws"
)

func main() {
	data, err := os.ReadFile("/tmp/portal_test.key")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	fs := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("fileserver got: %q\n", r.URL)
		http.FileServer(http.Dir(config.HTTPServer.FileRoot)).ServeHTTP(w, r)
	})
	r.Handle("/app/*", http.StripPrefix("/app/", fs))
	r.Handle("/ws/*", http.StripPrefix("/ws/", &ws.Server{

		Conns:    make(map[string]*websocket.Conn),
		Listener: any{},
	}))
	/*
		r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/app/index.html", http.StatusMovedPermanently)
		})
	*/

	c := cli.NewCommand("serve", "run http server", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
		fmt.Printf("listening at %q\n", config.HTTPServer.Addr)
		return http.ListenAndServe(config.HTTPServer.Addr, r)
	}), cli.WithHTTPServerFlags)

	err = c.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
