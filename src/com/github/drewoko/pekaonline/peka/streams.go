package peka

import (
	"gopkg.in/parnurzeal/gorequest.v0"
	"encoding/json"
)

func RequestStreams(pekaUserChannelRequest chan interface{})  {

	requester := gorequest.New()

	var request struct {
		Сontent string `json:"content"`
		Type string `json:"type"`
		Category struct{
			Slug string `json:"slug"`
		} `json:"category"`
	}

	request.Сontent = "stream"
	request.Type = "all"
	request.Category.Slug = "top"

	jsonText, _ := json.Marshal(request)

	_, body, _ := requester.
		Post("http://funstream.tv/api/content").
		Set("Accept", "application/json").
		Send(string(jsonText)).
		End()

	var response PekaStreamResponseStruct
	json.Unmarshal([]byte(body), &response)

	for _, content := range response.Content {
		if !content.TV {
			pekaUserChannelRequest <- content.Owner.Id
		}
	}
}

type PekaStreamResponseStruct struct{
	Content []struct{
		Owner struct{
			Id int `json:"id"`
		} `json:"owner"`
		TV bool `json:"tv"`
	} `json:"content"`
}
