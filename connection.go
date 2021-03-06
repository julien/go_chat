package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type connection struct {
	// Websocket connection
	ws *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Named props
	props map[string]string

	// A time stamp for the last message sent (avoid flooding)
	sentTime time.Time

	// Add "user managment"
	// banned map[string]*connection
}

func (c *connection) Reader() {
	for {
		_, msg, err := c.ws.ReadMessage()

		if err != nil {
			break
		}

		m := &message{connection: c, body: msg}
		h.broadcast <- m
	}
	//  c.ws.Close()
}

func (c *connection) Writer() {
	for msg := range c.send {

		ts := []byte(addTimeStamp())
		ms := [][]byte{ts, msg}
		output := bytes.Join(ms, []byte{' '})

		err := c.ws.WriteMessage(websocket.TextMessage, output)

		if err != nil {
			break
		}
	}
	// c.ws.Close()
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	c := &connection{
		send:     make(chan []byte, 256),
		ws:       ws,
		props:    make(map[string]string),
		sentTime: time.Now(),
	}
	h.register <- c

	defer func() {
		h.unregister <- c
	}()

	go c.Writer()
	c.Reader()
}
