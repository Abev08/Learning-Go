package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
)

// Simple HTTP server without external dependencies and packages.
// The server uses TCP connections and handles everything on it's own (no magic in the background).
// The server is single threaded and each connection (until it finishes) blocks other connections.
// It's enough for few connections and in my opinion you get more control over it.
// At the end the same functionality is created with build in HTTP server (for comparasion).

func main() {
	address := "127.0.0.1:8080"
	data, err := os.ReadFile("server/index.html")
	if err != nil {
		slog.Error("Error when reading index.html", "err", err)
		return
	}

	fmt.Printf("Starting HTTP server at: http://%s\n", address)

	// From scratch
	if true {
		var index = string(data)
		var tempBuff = make([]byte, 65535)
		var sb strings.Builder

		listener, err := net.Listen("tcp", address)
		if err != nil {
			slog.Error("Error when starting HTTP server", "err", err)
			return
		}

		for {
			var conn net.Conn
			var err error
			var req *http.Request
			sb.Reset()
			// .Accept() waits for new connection, it blocks code execution
			if conn, err = listener.Accept(); err != nil {
				fmt.Println(err)
				conn.Close()
				continue
			}

			n, err := conn.Read(tempBuff)
			if err != nil {
				slog.Error("Read error!", "Err", err)
			} else {
				// fmt.Print(string(readBuff[:n]))

				var reader = bufio.NewReader(bytes.NewReader(tempBuff[:n]))
				req, err = http.ReadRequest(reader)
				if err != nil {
					fmt.Println("Request read error", "Err", err)
					conn.Close()
					continue
				} else {
					// fmt.Print(req)
					sb.WriteString(fmt.Sprintf("New http %s request, url: %s", req.Method, req.URL))
					for _, v := range req.Header.Values("Upgrade") {
						sb.WriteString(fmt.Sprintf(", requested upgrade: %s", v))
					}
				}
			}

			// Check what was requested and create response
			switch req.URL.String() {
			case "/":
				conn.Write([]byte(fmt.Sprint("HTTP/1.1 200 OK\r\n",
					"Content-Length: ", len(index), "\r\n",
					"Content-Type: text/html\r\n\r\n",
					index,
					"\r\n\r\n")))
			default:
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}

			conn.Close()
			fmt.Println(sb.String())
		}
	} else {
		// Build in solution
		{
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write(data)
			})

			err = http.ListenAndServe(address, nil)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
