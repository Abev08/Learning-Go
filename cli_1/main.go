package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Simple CLI interface.
// The app counts and prints the result to console.
// Every 5 steps asks the user if it should continue.

func main() {
	reader := bufio.NewReader(os.Stdin)
	const sleep_dur = time.Millisecond * 1000 // Can also use time.Second
	i := 0
	for {
		i++
		fmt.Println(i)

		if i%5 == 0 {
			// Every 5 steps, ask the user if the app should continue
			fmt.Print("Continue? [y/n]")            // Ask the user
			var text, err = reader.ReadString('\n') // Read to the line break
			if err != nil {
				// Check if the read was successful
				fmt.Println(text, err)
			} else {
				// Trim whtie space characters and convert to lower case for comparasion
				text = strings.Trim(text, "\r\n ") // Leading and trailing '\r', '\n' and ' ' will be removed
				fmt.Printf("Received: %s\n", text)
				text = strings.ToLower(text)
				if text == "n" {
					break
				}
			}
		}

		time.Sleep(sleep_dur) // Slow down the loop
	}
}
