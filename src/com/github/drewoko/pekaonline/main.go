package main

import (

    GoodGame "./goodgame"
	Peka "./peka"
	Collector "./collector"
	Core "./core"
	"sync"
	"flag"
	"gopkg.in/magiconair/properties.v1"
)

func main() {

	configurationFile := flag.String("properties", "application.properties", "Properties file")
	flag.Parse()

	propertyFile := properties.MustLoadFile(*configurationFile, properties.UTF8)

	config := &Core.Config {
		Database: propertyFile.GetString("database", "db.db"),
		Port: propertyFile.GetString("port", "8080"),
		Static: propertyFile.GetString("static", "/static"),
		Dev: propertyFile.GetBool("dev", false),
	}

	goodgameMessageChan := make(chan bool)
	goodgameUserChannelRequest := make(chan interface{})
	goodgameUserChannelResponse := make(chan interface{})

	pekaMessageChan := make(chan bool)
	pekaUserChannelRequest := make(chan interface{})
	pekaUserChannelResponse := make(chan interface{})

	db := new(Core.DataBase).Init(config.Database)

	var wg sync.WaitGroup
	wg.Add(3)

	go GoodGame.InitChat(goodgameUserChannelRequest, goodgameUserChannelResponse, goodgameMessageChan)
	go Peka.InitChat(pekaUserChannelRequest, pekaUserChannelResponse, pekaMessageChan)

	go Collector.Start(goodgameUserChannelRequest, goodgameUserChannelResponse, pekaUserChannelRequest, pekaUserChannelResponse, goodgameMessageChan, pekaMessageChan, db)

	go Core.StartWeb(config, db)

	wg.Wait()
}