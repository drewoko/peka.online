package goodgame

import (
	"gopkg.in/parnurzeal/gorequest.v0"
	"encoding/json"
	"strconv"
)

func RequestStreams(goodgameUserChannelRequest chan interface{}) {
	requester := gorequest.New()

	_, body, _ := requester.
		Get("http://api2.goodgame.ru/v2/streams").
		Set("Accept", "application/vnd.goodgame.v2+json").
		End()

	var dat GGStreamResponseStruct
	json.Unmarshal([]byte(body), &dat)

	for i := 1; i <= dat.PageCount; i++ {

		_, body, _ := requester.
			Get("http://api2.goodgame.ru/v2/streams").
			Set("Accept", "application/vnd.goodgame.v2+json").
			Param("page", strconv.Itoa(i)).
			End()

		var dat GGStreamResponseStruct
		json.Unmarshal([]byte(body), &dat)

		for _, em := range dat.Embedded.Streams {
			goodgameUserChannelRequest <- em.Channel.Id
		}
	}
}

type GGStreamResponseStruct struct{
	PageCount int  `json:"page_count"`
	Page int  `json:"page"`

	Embedded struct{
		Streams []struct{
			Channel struct{
				Id interface{} `json:"id"`
			} `json:"channel"`
		}  `json:"streams"`
	} `json:"_embedded"`
}
