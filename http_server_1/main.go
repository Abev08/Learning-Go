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

func main() {
	fmt.Println("http://127.0.0.1:8080")
	data, err := os.ReadFile("server/index.html")
	if err != nil {
		slog.Error("Error when reading index.html", "Err", err)
		return
	}
	var index = string(data)
	var tempBuff = make([]byte, 65535)
	var sb strings.Builder

	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(listener)
	}

	for {
		var conn net.Conn
		var err error
		var req *http.Request
		sb.Reset()
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

	return

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Nothing"))
	})

	err = http.ListenAndServe("127.0.0.1:8080", nil)
	fmt.Println(err)
}
