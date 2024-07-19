package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

// Ngrok package allows to create TCP tunnels.
// It allows to create public url to access locally hosted app.
// The app uses code similar to http_server_1 example.

func main() {
	// Read token and domain address from secrets.txt file
	var token, domain string
	file, err := os.Open("secrets.txt")
	if err != nil {
		slog.Error("Error opening secrets file", "err", err)
		return
	}
	var reader = bufio.NewReader(file)
	var lineNum = 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		switch lineNum {
		case 0:
			token = strings.TrimSpace(line)
		case 1:
			domain = strings.TrimSpace(line)
		}
		lineNum++
	}
	file.Close()
	if len(token) > 0 && len(domain) > 0 {
		fmt.Println("Ngrok token and domain readed successfuly")
	}

	// Start ngrok tunnel
	tunnel, err := ngrok.Listen(context.Background(),
		// Random address
		config.HTTPEndpoint(),
		// Predefined address
		// config.HTTPEndpoint(config.WithDomain(domain)),
		ngrok.WithAuthtoken(token),
	)
	if err != nil {
		slog.Error("Error creating ngrok tunnel", "err", err)
		return
	}
	log.Println("Tunnel established at:", tunnel.URL())

	tempBuff := make([]byte, 65535)
	for {
		conn, err := tunnel.Accept()
		for {
			if err != nil {
				slog.Error("New connection error", "err", err)
				break
			}

			n, err := conn.Read(tempBuff)
			if err != nil {
				slog.Error("Connection read error", "err", err)
				break
			}
			req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(string(tempBuff[:n]))))
			if err != nil {
				slog.Error("HTTP parsing error", "err", err)
				break
			}

			slog.Info("New http request", "url", req.URL)
			// for h, v := range req.Header {
			// 	slog.Info("", "header", h, "value", v)
			// }

			switch req.URL.String() {
			case "/":
				conn.Write([]byte(fmt.Sprint("HTTP/1.1 200 OK\r\n",
					"Content-Length: 2\r\n",
					"Content-Type: text/html\r\n\r\n",
					"Hi",
					"\r\n\r\n")))
			default:
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}

			break
		}

		conn.Close()
	}
}
