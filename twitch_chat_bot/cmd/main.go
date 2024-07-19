package main

import (
	"time"
	"twitch_chat_bot/cmd/chat"
)

// Twitch chat bot.
// By default it runs on separate thread.
// It receives irc chat messages and parses them.
// Based on parsed message it can do different things:
// - respond with PONG message,
// - respond to predefined commands,
// - detect an event (subscription, raid, announcement, etc.),
// - send chat messages and responses to chat messages,
// The bot keeps queue of messages that should be sent, to not send them too often and exhaust the connection.
// Periodic messages can be easly implemented.

func main() {
	chat.Start()

	var sleepDur = time.Second
	for {
		time.Sleep(sleepDur)
	}
}
