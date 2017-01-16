package core

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"time"

	_ "gopkg.in/mattn/go-sqlite3.v1"
	"sync"
)

const AggregatePeriod int = 24 // in hours
const AggregatePatrs int = 24  // how many columns will be on graph

type RowMap map[string]interface{}

type DataBase struct {
	sync.Mutex
	db *sql.DB
}

type AggregateRow struct {
	Timing          int64
	SumUniqUsers    int
	SumMessageCount int
}

type AggregateStatistic struct {
	Goodgame []AggregateRow
	Peka     []AggregateRow
}

func getPeriodFrom() int64 {
	var period = time.Now().Add(-1 * time.Duration(AggregatePeriod) * time.Hour)
	var periodFrom int64 = time.Date(
		period.Year(),
		period.Month(),
		period.Day(),
		period.Hour(),
		0, 0, 0,
		period.Location(),
	).Unix()

	return periodFrom
}

func (d *DataBase) Init(path string) *DataBase {

	var err error
	d.db, err = sql.Open("sqlite3", path)

	if err != nil {
		log.Fatal("Failed to open database. Reason: ", err)
	}

	d.createTable()

	return d
}

func (d *DataBase) createTable() {
	d.db.Exec("CREATE TABLE statistics (id integer not null primary key, insertTime datetime default current_timestamp, externalTime text, uniqUsers int, userList text, messageCount int default 0, source text)")
}

func (d *DataBase) DeleteTemporaryTables(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	d.db.Exec("DROP TABLE goodgame_users_in_channels_cache" + id + "")
	d.db.Exec("DROP TABLE goodgame_message_cache" + id + "")
	d.db.Exec("DROP TABLE peka_users_in_channels_cache" + id + "")
	d.db.Exec("DROP TABLE peka_message_cache" + id + "")
}

func (d *DataBase) CreateTemporaryTables(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	var err error
	_, err = d.db.Exec("CREATE TABLE goodgame_users_in_channels_cache" + id + " (id integer not null primary key, requested text)")
	if err != nil {
		log.Println("CreateTemporaryTables1", err)
	}
	_, err = d.db.Exec("CREATE TABLE goodgame_message_cache" + id + " (id integer not null primary key)")
	if err != nil {
		log.Println("CreateTemporaryTables2", err)
	}
	_, err = d.db.Exec("CREATE TABLE peka_users_in_channels_cache" + id + " (id integer not null primary key, requested text)")
	if err != nil {
		log.Println("CreateTemporaryTables3", err)
	}
	_, err = d.db.Exec("CREATE TABLE peka_message_cache" + id + " (id integer not null primary key)")
	if err != nil {
		log.Println("CreateTemporaryTables4", err)
	}
}

func (d *DataBase) AddPekaMessageCache(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO peka_message_cache" + id + " (id) VALUES (null)")
	if err != nil {
		log.Println("AddPekaMessageCache", err)
	}
}

func (d *DataBase) InsertPekaUserChannelsCache(data string, id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO peka_users_in_channels_cache"+id+" (requested) VALUES (?)", data)
	if err != nil {
		log.Println("InsertPekaUserChannelsCache", err)
	}
}

func (d *DataBase) AddGGMessageCache(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO goodgame_message_cache" + id + " (id) VALUES (null)")
	if err != nil {
		log.Println("AddGGMessageCache", err)
	}
}

func (d *DataBase) GetGGMessagesCount(id string) int {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s := d.db.QueryRow("SELECT count(*) as cnt FROM goodgame_message_cache" + id)

	var cnt int
	s.Scan(&cnt)

	return cnt
}

func (d *DataBase) GetAllPekaUserInChannelsCache(id string) []string {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s, err := d.db.Query("SELECT requested FROM peka_users_in_channels_cache" + id)

	var caches []string
	if err == nil {
		for s.Next() {
			var requested string
			s.Scan(&requested)

			caches = append(caches, requested)
		}

	} else {
		log.Println("GetAllPekaUserInChannelsCache", err)
	}

	return caches
}

func (d *DataBase) InsertGGUserChannelsCache(data string, id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO goodgame_users_in_channels_cache"+id+" (requested) VALUES (?)", data)
	if err != nil {
		log.Println("InsertGGUserChannelsCache", err)
	}
}

func (d *DataBase) FlushGGUserChannelsCache() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("DELETE FROM goodgame_users_in_channels_cache")
	if err != nil {
		log.Println("FlushGGUserChannelsCache", err)
	}
}

func (d *DataBase) GetAllGoodGameUserInChannelsCache(id string) []string {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s, err := d.db.Query("SELECT requested FROM goodgame_users_in_channels_cache" + id)

	var caches []string
	if err == nil {

		for s.Next() {
			var requested string
			s.Scan(&requested)

			caches = append(caches, requested)
		}
	} else {
		log.Println("GetAllGoodGameUserInChannelsCache", err)
	}

	return caches
}

func (d *DataBase) GetPekaMessagesCount(id string) int {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s := d.db.QueryRow("SELECT count(*) as cnt FROM peka_message_cache" + id)

	var cnt int
	s.Scan(&cnt)

	return cnt
}

func (d *DataBase) InsertStatistics(externalTime string, uniqUsers int, userList string, messageCount int, source string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec(
		"INSERT INTO statistics (externalTime, uniqUsers, userList, messageCount, source) VALUES (?, ?, ?, ?, ?)",
		externalTime, uniqUsers, userList, messageCount, source)
	if err != nil {
		log.Println("InsertStatistics", err)
	}
}

func (d *DataBase) GetStatistics() []RowMap {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	sql := `
		SELECT 
			externalTime, 
			uniqUsers, 
			messageCount, 
			source
		
		FROM statistics 
		
		WHERE 
			externalTime != '' AND
			insertTime > datetime(?, 'unixepoch')

		ORDER BY externalTime DESC
	`

	s, _ := d.db.Query(sql, getPeriodFrom())
	defer s.Close()
	var rows []RowMap

	for s.Next() {
		row := make(RowMap)

		var externalTime string
		var uniqUsers int
		var messageCount int
		var source string

		s.Scan(&externalTime, &uniqUsers, &messageCount, &source)

		row["externalTime"] = externalTime
		row["uniqUsers"] = uniqUsers
		row["messageCount"] = messageCount
		row["source"] = source

		rows = append(rows, row)
	}

	return rows
}

func (d *DataBase) getAggregateStatisticsForSource(source string) []RowMap {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	sql := `
		SELECT 
			externalTime, 
			userList,
			messageCount
		
		FROM statistics 
		
		WHERE 
			externalTime != '' AND
			insertTime > datetime(?, 'unixepoch') AND
			source = ?

		ORDER BY externalTime ASC
	`

	s, err := d.db.Query(sql, getPeriodFrom(), source)
	var rows []RowMap
	if err != nil {
		log.Println("Aggregate SQL error", err)
		return rows
	}

	defer s.Close()

	for s.Next() {
		var timing string
		var userList string
		var sumMessageCount int

		s.Scan(&timing, &userList, &sumMessageCount)

		rows = append(rows, RowMap{
			"Timing":          timing,
			"UserList":        userList,
			"SumMessageCount": sumMessageCount,
		})
	}

	return rows
}

func aggregateResults(rows []RowMap) []AggregateRow {
	var period int64 = int64((AggregatePeriod / AggregatePatrs) * 60 * 60) // seconds
	var prevTiming int64 = 0

	var ARows []AggregateRow
	var ARow *AggregateRow = nil
	var usersGroup []string

	for _, row := range rows {
		timing, _ := strconv.ParseInt(row["Timing"].(string), 10, 64)

		if timing-prevTiming >= period {
			if ARow != nil {
				RemoveDuplicates(&usersGroup)
				ARow.SumUniqUsers = len(usersGroup)
				ARows = append(ARows, *ARow)

				usersGroup = usersGroup[:0]
			}

			prevTiming = timing
			ARow = &AggregateRow{0, 0, 0}
		}

		ARow.Timing = timing
		ARow.SumMessageCount += row["SumMessageCount"].(int)

		var users []string
		usersBytes := ([]byte)(row["UserList"].(string))

		err := json.Unmarshal(usersBytes, &users)
		if err != nil {
			log.Println("Can't json.Unmarshal user list", err)
			continue
		}

		usersGroup = append(usersGroup, users...)
	}

	if ARow != nil {
		RemoveDuplicates(&usersGroup)
		ARow.SumUniqUsers = len(usersGroup)

		ARows = append(ARows, *ARow)
	}

	return ARows
}

func (d *DataBase) GetAggregateStatistics() AggregateStatistic {
	goodGameStatistic := aggregateResults(d.getAggregateStatisticsForSource("goodgame"))
	pekaStatistic := aggregateResults(d.getAggregateStatisticsForSource("peka"))

	stat := AggregateStatistic{goodGameStatistic, pekaStatistic}

	return stat
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
