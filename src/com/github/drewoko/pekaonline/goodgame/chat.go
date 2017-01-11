package goodgame

import (
	"sync"
	"github.com/gorilla/websocket"
	"encoding/json"
	"log"
	"time"
)

type GoodGameSocketStorage struct {
	sync.Mutex
	wsClient *websocket.Conn
}

func (s *GoodGameSocketStorage) writeMessage(request []byte) error{
	s.Lock();
	defer s.Unlock()

	return s.wsClient.WriteMessage(websocket.TextMessage, request)
}

func InitChat(goodgameUserChannelRequest chan interface{}, goodgameUserChannelResponse chan interface{}, messageChan chan bool) {

	log.Println("Connecting to GoodGame.ru")

	wsClient, _, err := websocket.DefaultDialer.Dial("ws://chat.goodgame.ru:8081/chat/websocket", nil)

	if err != nil {
		log.Println("Failed to connect to GoodGame.ru", err)
		time.Sleep(time.Second * 5)
		InitChat(goodgameUserChannelRequest, goodgameUserChannelResponse, messageChan)
		return
	}

	socket := &GoodGameSocketStorage{
		wsClient: wsClient,
	}

	plainMessageChan := make(chan []byte)
	channelChan := make(chan string)
	quitChat := make(chan bool)

	pingTicker := time.NewTicker(time.Second * 5)

	defer func() {
		close(plainMessageChan)
		close(channelChan)
		pingTicker.Stop()
	}()

	go func() {
		for {
			messageType, message, err := wsClient.ReadMessage()

			if err != nil {
				log.Println("Disconnected from GoodGame.ru")
				quitChat <- true
				time.Sleep(time.Second * 5)
				InitChat(goodgameUserChannelRequest, goodgameUserChannelResponse, messageChan)
				return
			}

			if messageType == websocket.TextMessage {
				plainMessageChan <- message
			}
		}
	}()

	go func() {
		for {
			select {
				case cc := <- goodgameUserChannelRequest:
					joinToChannel(socket, cc)
					sentMessage(socket, GoodGameStruct{
						Type: "get_users_list2",
						Data: map[string]interface{}{"channel_id": cc},
					})
				case <-quitChat:
					break
			}
		}
	}()

	for {
		var plainMessage []byte
		select {
			case plainMessage = <- plainMessageChan:

				message := GoodGameStruct{}

				json.Unmarshal(plainMessage, &message)

				if message.Type == "welcome" {
					log.Println("Connected to GoodGame.ru")
				} else if message.Type == "users_list" {
					goodgameUserChannelResponse <- message.Data["users"]
				} else if message.Type == "message" {
					messageChan <- true
				}
			case <- pingTicker.C:
				sendPing(socket)
			case <-quitChat:
				break
		}
	}
}

func processChannels(counter *int, socket *GoodGameSocketStorage, message GoodGameStruct, channelChan chan string) {
	var channelInterface interface{}
	intCounter := 0;
	for _, channelInterface = range message.Data["channels"].([]interface{}) {
		channel := channelInterface.(map[string]interface{})["channel_id"].(string)
		joinToChannel(socket, channel)
		intCounter++
	}
	*counter = *counter + intCounter

	if intCounter == 50 {
		go requestChannels(socket, intCounter-1, 50)
	}
}

func sendPing(socket *GoodGameSocketStorage) {
	sentMessage(socket, GoodGameStruct{
		Type: "ping",
		Data: map[string]interface{}{},
	})
}

func joinToChannel(socket *GoodGameSocketStorage, channel interface{}) {
	sentMessage(socket, GoodGameStruct{
		Type: "join",
		Data: map[string]interface{}{"channel_id": channel, "hidden": false},
	})
}

func requestChannels(socket *GoodGameSocketStorage, start int, count int) {
	sentMessage(socket, GoodGameStruct{
		Type: "get_channels_list",
		Data: map[string]interface{}{"start": start, "count": count},
	})
}

func sentMessage(socket *GoodGameSocketStorage, messageStruct GoodGameStruct) {

	request, err := json.Marshal(messageStruct)

	if err != nil {
		log.Println("Failed to create JSON", err)
		return
	}

	err = socket.writeMessage(request)

	if err != nil {
		log.Println(err)
	}
}

type GoodGameStruct struct {
	Type string	`json:"type"`
	Data map[string]interface{} `json:"data"`
}