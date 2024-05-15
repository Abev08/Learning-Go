package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

const NGROK_AUTHTOKEN string = "xxx_yyy"
const NGROK_DOMAIN_ADDR string = "zzz.ngrok-free.app"

func main() {
	tunnel, err := ngrok.Listen(context.Background(),
		// Random address
		config.HTTPEndpoint(),
		// Predefined address
		// config.HTTPEndpoint(config.WithDomain(NGROK_DOMAIN_ADDR)),
		ngrok.WithAuthtoken(NGROK_AUTHTOKEN),
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
