package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type WebSocketConnection struct {
	Conn     *websocket.Conn
	Username string
}

type Message struct {
	From    string
	Message string
}

var connections = make([]*WebSocketConnection, 0)

func broadcastMessage(messageType int, message Message) {
	b, _ := json.Marshal(message)

	for _, eachConn := range connections {
		err := eachConn.Conn.WriteMessage(messageType, b)
		if err != nil {
			log.Println("write: ", err)
			break
		}
	}
}

func ejectConnection(currentConn *WebSocketConnection) {
	for i, eachConn := range connections {
		if currentConn == eachConn {
			connections = append(connections[:i], connections[i+1:]...)

			message := Message{
				From:    currentConn.Username,
				Message: " Left the Chat",
			}

			broadcastMessage(websocket.TextMessage, message)

			break
		}
	}
}

func main() {
	app := fiber.New()

	app.Static("/", "./index.html")

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		username := c.Query("username")

		currentConn := &WebSocketConnection{
			Conn:     c,
			Username: username,
		}

		connections = append(connections, currentConn)

		message := Message{
			From:    currentConn.Username,
			Message: " Joined the Chat",
		}

		broadcastMessage(websocket.TextMessage, message)

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				if strings.Contains(err.Error(), "websocket: close") {
					ejectConnection(currentConn)
				}

				log.Println("read: ", err)
				break
			}

			message := Message{
				From:    currentConn.Username,
				Message: string(msg),
			}

			broadcastMessage(mt, message)
		}
	}))

	log.Fatal(app.Listen(":3000"))
}
