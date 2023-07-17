package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type textMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func main() {
	var username string
	fmt.Print("Enter your username: ")
	fmt.Scanf("%s", &username)
	fmt.Printf("Hello, %v!\n", username)

	ctx := context.TODO()

	wsConn, _, err := websocket.Dial(ctx, "ws://localhost:8080/text", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	stopReceiving := make(chan struct{})
	go func() {
		var msg textMessage
		for {
			select {
			case <-stopReceiving:
				return
			default:
				err := wsjson.Read(context.TODO(), wsConn, &msg)
				if err != nil {
					log.Println(err)
					return
				}
				fmt.Println(msg.Username, msg.Content)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "exit" {
			close(stopReceiving)
			wsConn.Close(websocket.StatusNormalClosure, "Disconnecting")
			return
		}
		err := wsjson.Write(ctx, wsConn, &textMessage{Username: username, Content: line})
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
