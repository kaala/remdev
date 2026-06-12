package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"remdev/config"
	"remdev/handler"
	"remdev/pty"
)

//go:embed embed/*
var embedded embed.FS

func main() {
	cfg := config.Defaults()
	config.ParseFlags(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	ptyMgr := pty.NewManager()
	ptyMgr.WorkDir = cfg.Root

	dav, err := handler.NewWebDAV(cfg.Root)
	if err != nil {
		log.Fatalf("webdav: %v", err)
	}

	terminalHandler := handler.NewTerminalHandler(ptyMgr)
	auth := handler.Auth(cfg.Token)

	mux := http.NewServeMux()
	mux.Handle("/dav/", auth(dav))
	mux.Handle("/api/info", auth(handler.ServerInfoHandler()))
	mux.Handle("/ws/terminal", auth(terminalHandler))
	mux.Handle("/ws/terminal/", auth(terminalHandler))

	frontend, _ := embedded.ReadFile("embed/index.html")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(frontend)
			return
		}
		http.NotFound(w, r)
	})

	log.Printf("rdev listening on http://%s", addr)
	log.Printf("  root: %s", cfg.Root)
	if cfg.Token != "" {
		log.Printf("  token: enabled")
	}
	log.Fatal(http.ListenAndServe(addr, mux))
}
