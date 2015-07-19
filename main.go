package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
)

type Event struct {
    Operation string `json:"operation"`
    Channel string `json:"channel"`
    Data string `json:"data"`
}

var channels map[string]map[*websocket.Conn]bool

func socketHandler(ws *websocket.Conn) {
	
	fmt.Println("Connected")
	for {
		var event Event

		if err := websocket.JSON.Receive(ws, &event); err != nil {
			//remove this socket from all the channels
			for _, sockets := range channels {
				delete(sockets, ws)
			}

			fmt.Println("Error " + err.Error())
			return
		}

		fmt.Println("Received: ", event)

		switch event.Operation {
		case "subscribe":
			//channel binding
			if channels[event.Channel] == nil {
				fmt.Println("New Channel", event.Channel)
				channels[event.Channel] = make(map[*websocket.Conn]bool)
			}
			channels[event.Channel][ws] = true
		}
    }
}

func broadcast(event Event) {
	fmt.Println("Broadcast: ", event)
	for ws := range channels[event.Channel] {
		websocket.JSON.Send(ws, event)
	}
}

func pushHandler(w http.ResponseWriter,r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "post only", http.StatusMethodNotAllowed)
		return
	}
	//TODO: add user validation
	decoder := json.NewDecoder(r.Body)
	var event Event
	if err := decoder.Decode(&event); err != nil {
		http.Error(w,"Invalid Json", http.StatusBadRequest)
	}
	go broadcast(event)
}

const PORT=":8080"

func main() {
	channels = make(map[string]map[*websocket.Conn]bool)
	fmt.Println("Starting websocket server")
	
	//route uri
	http.Handle("/sock", websocket.Handler(socketHandler))
	http.HandleFunc("/push", pushHandler)

	if err := http.ListenAndServe(PORT, nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
