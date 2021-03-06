package main

import (
	"log"
	"strconv"
	"time"
)

type hub struct {
	// Registered connections
	connections map[*connection]bool

	// Register requests from connections
	register chan *connection

	// Unregister requests from connections
	unregister chan *connection

	// Inbound messages from connections
	broadcast chan *message
}

var h = hub{
	connections: make(map[*connection]bool),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	broadcast:   make(chan *message),
}

func (h *hub) NumConnections() int {
	count := 0
	for _, connected := range h.connections {
		if connected {
			count++
		}
	}
	return count
}

func (h *hub) ClientByName(name string) *connection {
	for c := range h.connections {
		if name == c.props["name"] {
			return c
		}
	}
	return nil
}

func (h *hub) ConnectionProperty(c *connection, prop string, value string) bool {
	if c.props[prop] != value {
		c.props[prop] = value
		return true
	} else {
		return false
	}
}

func (h *hub) Run() {
	for {
		select {

		case c := <-h.register:

			h.connections[c] = true

			currConn := strconv.FormatInt(int64(h.NumConnections()), 10)

			c.send <- []byte("Welcome")
			c.send <- []byte("")
			c.send <- []byte("***********************************")
			c.send <- []byte("Your user name is user" + currConn)
			c.send <- []byte("")
			c.send <- []byte("Type /name newname to change it")
			c.send <- []byte("Type /help for more")
			c.send <- []byte("***********************************")
			c.send <- []byte("")

			c.props["name"] = "user" + currConn
			c.props["id"] = string(currConn)

			for e := range h.connections {
				if e != c {
					select {
					case e.send <- []byte(c.props["name"] + " joined"):
					default:
						delete(h.connections, e)
						close(e.send)
					}
				}
			}

		case c := <-h.unregister:

			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}

		case m := <-h.broadcast:

			now := time.Now()
			diff := now.Sub(m.connection.sentTime)

			if diff.Seconds() < 1 {
				return
			}
			m.connection.sentTime = time.Now()

			isCmd, name, args := m.ToCommand()
			if isCmd {

				switch name {
				case "name":
					oldName := m.connection.props["name"]
					changed := h.ConnectionProperty(m.connection, "name", args[0])
					if changed {
						// m.connection.send <- []byte("Your name was changed to" + m.connection.props["name"])

						for c := range h.connections {
							select {
							case c.send <- []byte(oldName + " changed his name to " + m.connection.props["name"]):
							}
						}
					}

				case "users":
					m.connection.send <- []byte("")
					m.connection.send <- []byte("***********************************")
					m.connection.send <- []byte("Connected users:")

					// names := make([]string, h.NumConnections())
					for c := range h.connections {
						m.connection.send <- []byte(" + " + c.props["name"])
					}
					m.connection.send <- []byte("***********************************")
					m.connection.send <- []byte("")

				default:
					m.connection.send <- []byte("")
					m.connection.send <- []byte("***********************************")
					m.connection.send <- []byte("Chat help")
					m.connection.send <- []byte("")
					m.connection.send <- []byte("")
					m.connection.send <- []byte("***********************************")
					m.connection.send <- []byte("/name newname: Set your user name")
					m.connection.send <- []byte("/help: Display this message")
					m.connection.send <- []byte("***********************************")
					m.connection.send <- []byte("")
				}

			} else {

				msg := []byte(m.connection.props["name"] + " : " + string(m.body))

				for c := range h.connections {
					select {
					case c.send <- msg:
					default:
						log.Printf("Got %q\n", c)
						// delete(h.connections, c)
						// close(c.send)
					}
				}
			}

		}
	}
}
