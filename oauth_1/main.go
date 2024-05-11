package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"
)

const TWITCH_CLIENT_ID string = ""   // Twitch API bots client ID
const TWITCH_CLIENT_PASS string = "" // Twitch API bots client password
const TWITCH_REDIRECT_URI string = "http://localhost:3000"         // Twitch API bots redirect Uri
var TWITCH_SCOPES []string = []string{
	"bits:read",                     // View Bits information for a channel
	"channel:manage:redemptions",    // Manage Channel Points custom rewards and their redemptions on a channel
	"channel:moderate",              // Perform moderation actions in a channel. The user requesting the scope must be a moderator in the channel
	"channel:read:hype_train",       // View Hype Train information for a channel
	"channel:read:redemptions",      // View Channel Points custom rewards and their redemptions on a channel
	"channel:read:subscriptions",    // View a list of all subscribers to a channel and check if a user is subscribed to a channel
	"chat:edit",                     // Send live stream chat messages
	"chat:read",                     // View live stream chat messages
	"moderator:manage:banned_users", // Ban and unban users
	"moderator:manage:shoutouts",    // Manage a broadcaster’s shoutouts
	"moderator:read:chatters",       // View the chatters in a broadcaster’s chat room
	"moderator:read:followers",      // Read the followers of a broadcaster
	"whispers:edit",                 // Send whisper messages
	"whispers:read",                 // View your whisper messages
}

var TwitchToken, TwitchTokenRefresh string
var TwitchTokenExpirationDate time.Time

func main() {
	// OAuth validation based on Twitch API
	// The TwitchToken, TwitchTokenRefresh, TwitchTokenExpirationDate variables should be saved to file / database
	// after token refresh the file / database should be updated.
	// Using previous Token and TokenRefresh variables allows to update the token without user interference.

	if len(TwitchToken) == 0 || len(TwitchTokenRefresh) == 0 {
		getNewTwitchToken()
	} else {
		if !validateTwitchToken() {
			if !refreshTwitchToken() {
				getNewTwitchToken()
			}
		}
	}
}

// Requests new Twitch access token
func getNewTwitchToken() {
	slog.Info("Twitch access token, requesting new one.")
	var err error
	var code string

	var url = fmt.Sprintf(
		"https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		TWITCH_CLIENT_ID,
		TWITCH_REDIRECT_URI,
		// url.QueryEscape(strings.Join(TWITCH_SCOPES, "+")), // Query escape also escapes "+" which Twitch doesn't like
		strings.ReplaceAll(strings.Join(TWITCH_SCOPES, "+"), ":", "%3A"),
	)

	// Open the url for the user to complete authorization
	if err = openUrl(url); err != nil {
		slog.Error("Twitch access token request. Error when opening url", "Err", err)
		return
	}

	// Local server is needed to get response to user authorizing the app (to grab the access token)
	var server net.Listener
	server, err = net.Listen("tcp", "localhost:3000")
	if err != nil {
		slog.Error("Twitch access token request. Couldn't start local server", "Err", err)
		return
	}
	for {
		var conn net.Conn
		conn, err = server.Accept()
		if err != nil {
			slog.Error("Twitch access token request. Error accepting new connetion to local server", "Err", err)
			continue
		}

		var tempBuff = make([]byte, 65535)
		var n int
		n, err = conn.Read(tempBuff)
		if err != nil {
			conn.Close()
			slog.Error("Twitch access token request. Couldn't read data from accepted local server connection", "Err", err)
			continue
		}
		var req *http.Request
		req, err = http.ReadRequest(bufio.NewReader(bytes.NewReader(tempBuff[:n])))
		if err != nil {
			conn.Close()
			slog.Error("Twitch access token request. Couldn't parse received request", "Err", err)
			continue
		}

		for query, value := range req.URL.Query() {
			switch query {
			case "code":
				if len(value) > 0 {
					code = value[0]
				}
			case "scope":
				// Check if got permissions for all of the requested scopes
				if len(value) > 0 && len(TWITCH_SCOPES) != (strings.Count(value[0], " ")+1) {
					slog.Warn("Twitch access token request. Received request didn't support all of the requested scopes")
				}
			}
		}

		// Redirect to www.twitch.tv to hide code part in the url
		conn.Write([]byte(fmt.Sprint("HTTP/1.1 302 Found\r\n",
			"Location: https://www.twitch.tv\r\n\r\n")))
		conn.Close()

		if len(code) > 0 {
			break // Received code part, continue
		} else {
			slog.Warn("Twitch access token request. Received request doesn't contain code part - waiting for another connection")
		}
	}

	// Close local server, it's no longer needed
	server.Close()

	// Next step - request user token with received authorization code
	var resp *http.Response
	resp, err = http.Post("https://id.twitch.tv/oauth2/token", "application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=%s",
			TWITCH_CLIENT_ID,
			TWITCH_CLIENT_PASS,
			code,
			TWITCH_REDIRECT_URI,
		)),
	)
	if err != nil {
		slog.Error("Twitch access token request. Error when requesting OAuth token with received auth code.", "Err", err)
		return
	}
	if resp.StatusCode != 200 {
		slog.Error("Twitch access token request. Request didn't succeed", "Code", resp.StatusCode)
		return
	}

	var reader = bufio.NewReader(resp.Body)
	var data []byte
	for {
		data, err = reader.ReadBytes('\n')
		if err != nil {
			break
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(data, &responseMap)
		if err != nil {
			slog.Error("Twitch access token request. Error when parsing json response", "Err", err)
			continue
		}

		// Get the data
		TwitchToken = responseMap["access_token"].(string)
		TwitchTokenRefresh = responseMap["refresh_token"].(string)
		var expiresIn = responseMap["expires_in"].(float64)
		// Check the scopes
		var scopes = responseMap["scope"].([]interface{})
		if !checkTwitchScopes(scopes) {
			slog.Error("Twitch access token request. Received scopes are different from requested ones")
		}

		var expirationDuration = time.Second * time.Duration(expiresIn)
		TwitchTokenExpirationDate = time.Now().Add(expirationDuration)
		slog.Info(fmt.Sprintf("Twitch access token request successful. Token expires in %d seconds (%s)",
			int(expiresIn),
			expirationDuration.String()))
	}
}

// Validates Twitch access token. Returns true if validation was successful, otherwise false.
func validateTwitchToken() bool {
	slog.Info("Twitch access token validation started.")
	if len(TWITCH_CLIENT_ID) == 0 || len(TwitchToken) == 0 {
		slog.Warn("Twitch access token validation failed. Missing Client ID or OAuth token.")
		return false
	}

	var err error
	var req *http.Request
	var resp *http.Response
	req, err = http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		slog.Error("Twitch access token validation. Error when creating request", "Err", err)
		return false
	}
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", TwitchToken))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Twitch access token validation. Error when sending request", "Err", err)
		return false
	}
	if resp.StatusCode != 200 {
		slog.Error("Twitch access token validation. Request didn't succeed", "Status", resp.Status)
		return false
	}

	var reader = bufio.NewReader(resp.Body)
	var data []byte
	for {
		data, err = reader.ReadBytes('\n')
		if err != nil {
			break
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(data, &responseMap)
		if err != nil {
			slog.Error("Twitch access token validation. Error when parsing json response", "Err", err)
			continue
		}

		// Get the data
		var clientID = responseMap["client_id"].(string)
		var expiresIn = responseMap["expires_in"].(float64)
		if clientID != TWITCH_CLIENT_ID || expiresIn <= 0 {
			slog.Error("Twitch access token validation. Token has expired")
			return false
		}
		// Update expiration date
		var expirationDuration = time.Second * time.Duration(expiresIn)
		TwitchTokenExpirationDate = time.Now().Add(expirationDuration)

		// Check the scopes
		var scopes = responseMap["scopes"].([]interface{})
		if !checkTwitchScopes(scopes) {
			slog.Error("Twitch access token validation. Received scopes are different from requested ones")
			return false
		}

		slog.Info(fmt.Sprintf("Twitch access token validation successful. Token expires in %d seconds (%s)",
			int(expiresIn),
			expirationDuration.String()))
		return true
	}

	return false
}

// Refreshes Twitch access token. Returns true if refresh was successful, otherwise false.
func refreshTwitchToken() bool {
	slog.Info("Twitch access token refresh started.")
	if len(TWITCH_CLIENT_ID) == 0 || len(TWITCH_CLIENT_PASS) == 0 || len(TwitchTokenRefresh) == 0 {
		slog.Warn("Twitch access token refresh failed. Missing Client ID, Client Password or OAuth refresh token.")
		return false
	}

	var err error
	var resp *http.Response
	resp, err = http.Post("https://id.twitch.tv/oauth2/token", "application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s",
			TWITCH_CLIENT_ID,
			TWITCH_CLIENT_PASS,
			url.QueryEscape(TwitchTokenRefresh),
		)),
	)
	if err != nil {
		slog.Error("Twitch access token refresh. Error when sending request.", "Err", err)
		return false
	}
	if resp.StatusCode != 200 {
		slog.Error("Twitch access token refresh. Request didn't succeed", "Code", resp.StatusCode)
		return false
	}

	var reader = bufio.NewReader(resp.Body)
	var data []byte
	for {
		data, err = reader.ReadBytes('\n')
		if err != nil {
			break
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(data, &responseMap)
		if err != nil {
			slog.Error("Twitch access token refresh. Error when parsing json response", "Err", err)
			continue
		}

		// Get the data
		TwitchToken = responseMap["access_token"].(string)
		TwitchTokenRefresh = responseMap["refresh_token"].(string)
		var expiresIn = responseMap["expires_in"].(float64)
		// Update expiration date
		var expirationDuration = time.Second * time.Duration(expiresIn)
		TwitchTokenExpirationDate = time.Now().Add(expirationDuration)

		// Check the scopes
		var scopes = responseMap["scope"].([]interface{})
		if !checkTwitchScopes(scopes) {
			slog.Error("Twitch access token refresh. Received scopes are different from requested ones")
			return false
		}

		slog.Info(fmt.Sprintf("Twitch access token refresh successful. Token expires in %d seconds (%s)",
			int(expiresIn),
			expirationDuration.String()))
		return true
	}

	return false
}

// Compares provided list of scopes with TWITCH_SCOPES returning true if they are the same, otherwise false.
func checkTwitchScopes(scopes []interface{}) bool {
	if len(scopes) != len(TWITCH_SCOPES) {
		return false
	}

	// Length is the same, check each scope
	var requestedScopes = make([]string, len(TWITCH_SCOPES))
	copy(requestedScopes, TWITCH_SCOPES)
	for _, v := range scopes {
		var idx = slices.Index(TWITCH_SCOPES, v.(string))
		if idx == -1 {
			return false
		}
		requestedScopes[idx] = ""
	}
	for _, v := range requestedScopes {
		if len(v) != 0 {
			return false
		}
	}
	return true
}

func openUrl(url string) error {
	var err error = nil

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}
