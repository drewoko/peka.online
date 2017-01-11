package peka

import (
	"log"
	"time"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"strconv"
)

func InitChat(pekaUserChannelRequest chan interface{}, pekaUserChannelResponse chan interface{},  messageChan chan bool) {

	log.Println("Trying to connect to Funstream.tv WS")

	ws_client, err := gosocketio.Dial(
		gosocketio.GetUrl("chat.funstream.tv", 80, false),
		&transport.WebsocketTransport{
			PingInterval:   5 * time.Second,
			PingTimeout:    10 * time.Second,
			ReceiveTimeout: 10 * time.Second,
			SendTimeout:    10 * time.Second,
			BufferSize:     1024 * 32,
		})

	if err != nil {
		log.Println("Failed to connect to funstreams.tv WS. Reason: ", err)

		time.Sleep(time.Second * 10)
		InitChat(pekaUserChannelRequest, pekaUserChannelResponse, messageChan)
		return
	}

	quitChan := make(chan bool)

	ws_client.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Funstream.tv WS connected")
	})

	ws_client.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Println("Disconnected from Peka2.tv")
		quitChan <- true
	})

	ws_client.On("/chat/message", func(h *gosocketio.Channel, args Message) {
		messageChan <- true
	})

	for {
		select {
		case cc := <- pekaUserChannelRequest:
			go func() {
				channel := "stream/"+strconv.Itoa(cc.(int))

				ws_client.Emit("/chat/join", struct {
					Channel string `json:"channel"`
				} { channel })

				res, err := ws_client.Ack("/chat/channel/list", struct {
					Channel string `json:"channel"`
				} { Channel: channel }, time.Second*10)

				if err == nil {
					pekaUserChannelResponse <- res
				}
			}()
		case <-quitChan:
			close(quitChan)
			InitChat(pekaUserChannelRequest, pekaUserChannelResponse, messageChan)
			return
		}
	}
}

type Message struct {
	Id      int64    `json:"id"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
	From	struct {
		Name	string `json:"name"`
	} `json:"from"`
}