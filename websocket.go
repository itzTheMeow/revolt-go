package revoltgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sacOO7/gowebsocket"
)

func (c *Client) Start() {
	// Create new socket
	c.Socket = gowebsocket.New(WS_URL)
	c.HTTP = &http.Client{}

	// Send auth when connected
	c.Socket.OnConnected = func(_ gowebsocket.Socket) {
		c.Socket.SendText(fmt.Sprintf("{\"type\": \"Authenticate\", \"token\": \"%s\"}", c.Token))
	}

	c.Socket.OnTextMessage = func(message string, _ gowebsocket.Socket) {
		// Parse data
		rawData := &struct {
			Type string `json:"type"`
		}{}
		err := json.Unmarshal([]byte(message), rawData)

		if err != nil {
			c.Destroy()
			panic(err)
		}

		if rawData.Type == "Authenticated" {
			go c.ping()
		}

		// Check events
		if rawData.Type == "Ready" && c.OnReadyFunction != nil {
			// Ready Event
			c.OnReadyFunction()
		} else if rawData.Type == "Message" && c.OnMessageFunction != nil {
			// Message Event
			msgData := &Message{}
			err := json.Unmarshal([]byte(message), msgData)

			if err != nil {
				fmt.Printf("Unexcepted Error: %s", err)
			}

			msgData.Client = c
			c.OnMessageFunction(msgData)
		} else if rawData.Type == "MessageUpdate" && c.OnMessageUpdateFunction != nil {
			// Message Update Event
			data := &struct {
				ChannelId string                 `json:"channel"`
				MessageId string                 `json:"id"`
				Payload   map[string]interface{} `json:"data"`
			}{}

			err := json.Unmarshal([]byte(message), data)

			if err != nil {
				fmt.Printf("Unexcepted Error: %s", err)
			}

			c.OnMessageUpdateFunction(data.ChannelId, data.MessageId, data.Payload)
		} else if rawData.Type == "MessageDelete" && c.OnMessageDeleteFunction != nil {
			// Message Delete Event
			data := &struct {
				ChannelId string `json:"channel"`
				MessageId string `json:"id"`
			}{}

			err := json.Unmarshal([]byte(message), data)

			if err != nil {
				fmt.Printf("Unexcepted Error: %s", err)
			}

			c.OnMessageDeleteFunction(data.ChannelId, data.MessageId)
		}

		fmt.Println(message)
	}

	// Start connection
	c.Socket.Connect()
}

// Destroy the websocket.
func (c *Client) Destroy() {
	c.Socket.Close()
}

// Ping websocket.
func (c *Client) ping() {
	for {
		time.Sleep(30 * time.Second)
		c.Socket.SendText(fmt.Sprintf("{\"type\":\"Ping\",\"time\":%d}", time.Now().Unix()))
	}
}
