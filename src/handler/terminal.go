package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"remdev/pty"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type TerminalHandler struct {
	manager *pty.Manager
}

func NewTerminalHandler(mgr *pty.Manager) *TerminalHandler {
	return &TerminalHandler{manager: mgr}
}

func (h *TerminalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uuidStr := r.URL.Query().Get("uuid")

	if uuidStr == "" {
		newUUID := genUUID()
		http.Redirect(w, r, "/ws/terminal?uuid="+newUUID, http.StatusFound)
		return
	}

	// The browser's fetch() follows the redirect with a plain HTTP GET.
	// Only perform the WebSocket upgrade when the Upgrade header is present.
	if r.Header.Get("Upgrade") != "websocket" {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.handleWS(w, r, uuidStr)
}

type wsMsg struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
	Code int    `json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}

func (h *TerminalHandler) handleWS(w http.ResponseWriter, r *http.Request, id string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Wait for the first resize message before creating the PTY,
	// so the shell starts with the correct terminal dimensions.
	cols, rows := waitForResize(conn)
	if cols == 0 {
		cols, rows = 80, 24
	}

	term, err := h.manager.CreateWithSize(id, cols, rows)
	if err != nil {
		sendMsg(conn, wsMsg{Type: "error", Text: err.Error()})
		return
	}
	defer h.manager.Remove(id)

	var wg sync.WaitGroup
	wg.Add(2)

	// PTY → WebSocket
	go func() {
		defer wg.Done()
		defer conn.Close()
		buf := make([]byte, 4096)
		for {
			n, err := term.Read(buf)
			if n > 0 {
				data := base64.StdEncoding.EncodeToString(buf[:n])
				if sendErr := sendMsg(conn, wsMsg{Type: "output", Data: data}); sendErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	sendMsg(conn, wsMsg{Type: "title", Text: term.Title()})

	// WebSocket → PTY
	go func() {
		defer wg.Done()
		defer term.Kill()
		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var msg wsMsg
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "input":
				data, err := base64.StdEncoding.DecodeString(msg.Data)
				if err != nil {
					continue
				}
				term.Write(data)
			case "resize":
				if msg.Cols > 0 && msg.Rows > 0 {
					term.Resize(msg.Cols, msg.Rows)
				}
			}
		}
	}()

	wg.Wait()

	code := term.Wait()
	sendMsg(conn, wsMsg{Type: "exit", Code: code})
}

func waitForResize(conn *websocket.Conn) (cols, rows int) {
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			return 0, 0
		}
		var msg wsMsg
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		if msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
			return msg.Cols, msg.Rows
		}
	}
}

func sendMsg(conn *websocket.Conn, msg wsMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func genUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
