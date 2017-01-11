package collector

import (
	"time"
	Core "../core"
	GoodGame "../goodgame"
	Peka "../peka"
	"encoding/json"
	"strconv"
	"log"
)

var (
	prevId string = ""
	curId string
)

func Start(
	goodgameUserChannelRequest chan interface{},
	goodgameUserChannelResponse chan interface{},
	pekaUserChannelRequest chan interface{},
	pekaUserChannelResponse chan interface{},
	goodgameMessageChan chan bool,
	pekaMessageChan chan bool,
	db *Core.DataBase) {

	fiveMinTicker := time.NewTicker(time.Minute)

	go func() {
		for ;; {
			select {
				case <- goodgameMessageChan:
					db.AddGGMessageCache(curId)
				case <- pekaMessageChan:
					db.AddPekaMessageCache(curId)
				case ggResp := <- goodgameUserChannelResponse:
					jsonString, err := json.Marshal(ggResp)
					if err == nil {
						db.InsertGGUserChannelsCache(string(jsonString), curId)
					}
				case pekaResp := <-pekaUserChannelResponse:
					db.InsertPekaUserChannelsCache(pekaResp.(string), curId)
			}
		}
	}()

	tick(goodgameUserChannelRequest, pekaUserChannelRequest, db)

	for {
		select {
			case <-fiveMinTicker.C:
				tick(goodgameUserChannelRequest, pekaUserChannelRequest, db)
		}
	}
}

func tick(goodgameUserChannelRequest chan interface{}, pekaUserChannelRequest chan interface{}, db *Core.DataBase) {

	startTime := time.Now().Unix()
	startTimeString := strconv.FormatInt(startTime, 10)

	log.Println("Starting tick", startTimeString)

	prevId = curId
	curId = startTimeString

	db.CreateTemporaryTables(startTimeString)

	Peka.RequestStreams(pekaUserChannelRequest)
	GoodGame.RequestStreams(goodgameUserChannelRequest)
	go func() {
		curIdToProcess := prevId

		log.Println("Starting post processing")
		postProcess(db, curIdToProcess)

		db.DeleteTemporaryTables(prevId)
	}()
}

func postProcess(db *Core.DataBase, curIdToProcess string) {

	goodgamePostProcess(db, curIdToProcess)
	pekaPostProcess(db, curIdToProcess)
}

func pekaPostProcess(db *Core.DataBase, curIdToProcess string)  {
	var PekaUsers []string

	for _, cache := range db.GetAllPekaUserInChannelsCache(curIdToProcess) {

		var pekaCacheStruct struct{
			Status string  `json:"status"`
			Result struct{
				Users []struct{
					Name string  `json:"name"`
				}  `json:"users"`
			}  `json:"result"`
		}

		json.Unmarshal([]byte(cache), &pekaCacheStruct)

		if pekaCacheStruct.Status == "ok" {
			for _, user := range pekaCacheStruct.Result.Users {
				PekaUsers = append(PekaUsers, user.Name)
			}
		}
	}

	RemoveDuplicates(&PekaUsers)

	jsonPekaUsers, _ := json.Marshal(PekaUsers)

	db.InsertStatistics(curIdToProcess, len(PekaUsers), string(jsonPekaUsers), db.GetPekaMessagesCount(curIdToProcess),"peka")
}

func goodgamePostProcess(db *Core.DataBase, curIdToProcess string) {
	var GGUsers []string

	for _, cache := range db.GetAllGoodGameUserInChannelsCache(curIdToProcess) {

		var GGCacheStruct []struct{
			Name string  `json:"name"`
		}

		json.Unmarshal([]byte(cache), &GGCacheStruct)

		for _, user := range GGCacheStruct {
			GGUsers = append(GGUsers, user.Name)
		}
	}

	RemoveDuplicates(&GGUsers)

	jsonGGUsers, _ := json.Marshal(GGUsers)

	db.InsertStatistics(curIdToProcess, len(GGUsers), string(jsonGGUsers), db.GetGGMessagesCount(curIdToProcess), "goodgame")

}

func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}