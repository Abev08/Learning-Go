package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{}
var wsConns = make([]*WSConn, 0, 10)
var wsNewConns = make(chan *WSConn, 10)

func main() {
	go handleWebSocketConns()

	ip := "127.0.0.1"
	port := 80

	http.HandleFunc("/", handleHTTPRequest)
	address := fmt.Sprintf("%s:%v", ip, port)
	Log("INF", "HTTP server, started at: http://", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	Log("INF", "HTTP server, new request from: ", r.RemoteAddr, ", ", r.Method, " ", r.RequestURI, " ", r.Proto)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.RequestURI {
	case "/":
		if websocket.IsWebSocketUpgrade(r) {
			ws, err := wsUpgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			wsNewConns <- NewWSConn(ws)
		} else {
			sendFile("www/client.html", w)
		}
	case "/client.js":
		sendFile("www/client.js", w)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func sendFile(filePath string, w http.ResponseWriter) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Write(data)
}

func handleWebSocketConns() {
	for {
		if len(wsNewConns) != 0 {
			ws := <-wsNewConns
			if ws != nil {
				added := false
				// Check nil spots
				for i := range wsConns {
					if wsConns[i] == nil {
						wsConns[i] = ws
						added = true
					}
				}
				if !added {
					wsConns = append(wsConns, ws)
				}
			}
		}

		for i := len(wsConns) - 1; i >= 0; i-- {
			conn := wsConns[i]
			if conn == nil {
				continue
			}

			if conn.shouldClose {
				conn.ws.Close()
				wsConns[i] = nil
				continue
			}

			// Read
			if !conn.readActive {
				conn.ReadAsync()
			}

			// Write
			if conn.sendRequested {
				conn.sendRequested = false
				conn.ws.WriteMessage(websocket.TextMessage, conn.sendBuffer)
			}
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func Log(level string, args ...string) {
	sb := strings.Builder{}
	sb.WriteRune('[')
	sb.WriteString(time.Now().Format("15:04:05.000"))
	sb.WriteRune(' ')
	sb.WriteString(level)
	sb.WriteString("] ")
	for i := range args {
		sb.WriteString(args[i])
	}
	fmt.Println(sb.String())
}

type WSConn struct {
	ws            *websocket.Conn
	readActive    bool
	sendRequested bool
	sendBuffer    []byte
	shouldClose   bool
	remoteAddr    string
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{
		ws:         conn,
		remoteAddr: conn.RemoteAddr().String(),
	}
}

func (conn *WSConn) ReadAsync() {
	conn.readActive = true
	go func() {
		messageType, data, err := conn.ws.ReadMessage()
		conn.readActive = false

		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				conn.shouldClose = true
				Log("ERR", "WebSocket connection from ", conn.remoteAddr, " closed")
			default:
				Log("ERR", "WebSocket connection from ", conn.remoteAddr, " error: ", err.Error())
			}
			return
		}

		switch messageType {
		case websocket.TextMessage:
			msg := string(data)
			Log("INF", "WebSocket connection from ", conn.remoteAddr, " received message: ", msg)

			switch msg {
			case "PING":
				conn.sendBuffer = []byte("PONG")
				conn.sendRequested = true
			case "Hello?":
				conn.sendBuffer, err = json.Marshal(map[string]string{"message": "Hi!"})
				if err == nil {
					conn.sendRequested = true
				}
			}
		default:
			Log("WRN", "WebSocket connection from ", conn.remoteAddr, " received unsupported message type: ", strconv.Itoa(messageType))
		}
	}()
}
