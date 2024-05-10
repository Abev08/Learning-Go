package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	var i int32 = 0
	const sleep_dur = time.Millisecond * 1234
	reader := bufio.NewReader(os.Stdin)
	for {
		time.Sleep(sleep_dur)
		fmt.Println(i)
		i++

		if i%5 == 0 {
			fmt.Print("Continue?")
			var text, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println(text, err)
			} else {
				var s string
				_ =
					strings.Trim(text, s)
				if s == "0" {
					break
				}
				fmt.Printf("Received: %s\n", text)
			}
		}
	}
}
