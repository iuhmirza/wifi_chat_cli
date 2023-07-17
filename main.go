package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/mdns"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var servers []*mdns.ServiceEntry

type textMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func main() {
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	mdns.Lookup("_wifichat._tcp", entriesCh)

	go func() {
		count := 0
		for entry := range entriesCh {
			servers = append(servers, entry)
			fmt.Println(count, " ", entry.Info)
			count += 1
		}
	}()

	var choice int
	fmt.Println("Enter number to connect to server")
	_, err := fmt.Scanf("%d", &choice)
	close(entriesCh)
	if err != nil {
		log.Println(err)
		return
	}
	if choice < 0 || len(servers) <= choice {
		log.Println("choice out of range")
		return
	}

	var username string
	fmt.Print("Enter your username: ")
	fmt.Scanf("%s", &username)
	fmt.Printf("Hello, %v!\n", username)

	ctx := context.TODO()

	wsConn, _, err := websocket.Dial(
		ctx,
		fmt.Sprintf("ws://%v:%v/text", servers[choice].Addr, servers[choice].Port),
		nil)
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
				err = wsjson.Read(context.TODO(), wsConn, &msg)
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
		err = wsjson.Write(ctx, wsConn, &textMessage{Username: username, Content: line})
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
