package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"remdev/config"
	"remdev/handler"
	"remdev/pty"
)

//go:embed embed/*
var embedded embed.FS

func main() {
	cfg := config.Defaults()
	config.ParseFlagsAndMerge(cfg)

	// Create default config file on first run.
	if _, err := os.Stat(cfg.ConfigPath()); os.IsNotExist(err) {
		if err := cfg.Save(); err != nil {
			log.Printf("warning: could not create config: %v", err)
		}
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	ptyMgr := pty.NewManager()

	dav, err := handler.NewWebDAV(cfg.Root)
	if err != nil {
		log.Fatalf("webdav: %v", err)
	}

	configHandler := handler.NewConfigHandler(cfg)
	terminalHandler := handler.NewTerminalHandler(ptyMgr)
	auth := handler.Auth(cfg.Token)

	mux := http.NewServeMux()
	mux.Handle("/dav/", auth(dav))
	mux.Handle("/api/config", auth(configHandler))
	mux.Handle("/api/serverinfo", auth(handler.ServerInfoHandler()))
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
	log.Printf("  config: %s", cfg.ConfigPath())
	if cfg.Token != "" {
		log.Printf("  token: enabled")
	}
	log.Fatal(http.ListenAndServe(addr, mux))
}
