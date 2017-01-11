package core

import (
	"log"
	"database/sql"

	_ "gopkg.in/mattn/go-sqlite3.v1"
	"sync"
)

type RowMap map[string]interface{}

type DataBase struct {
	sync.Mutex
	db *sql.DB
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
	d.db.Exec("DROP TABLE goodgame_users_in_channels_cache"+id+"")
	d.db.Exec("DROP TABLE goodgame_message_cache"+id+"")
	d.db.Exec("DROP TABLE peka_users_in_channels_cache"+id+"")
	d.db.Exec("DROP TABLE peka_message_cache"+id+"")
}

func (d *DataBase) CreateTemporaryTables(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	var err error
	_,err=d.db.Exec("CREATE TABLE goodgame_users_in_channels_cache"+id+" (id integer not null primary key, requested text)")
	if err != nil {log.Println("CreateTemporaryTables1", err)}
	_,err=d.db.Exec("CREATE TABLE goodgame_message_cache"+id+" (id integer not null primary key)")
	if err != nil {log.Println("CreateTemporaryTables2", err)}
	_,err=d.db.Exec("CREATE TABLE peka_users_in_channels_cache"+id+" (id integer not null primary key, requested text)")
	if err != nil {log.Println("CreateTemporaryTables3", err)}
	_,err=d.db.Exec("CREATE TABLE peka_message_cache"+id+" (id integer not null primary key)")
	if err != nil {log.Println("CreateTemporaryTables4", err)}
}

func (d *DataBase) AddPekaMessageCache(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO peka_message_cache"+id+" (id) VALUES (null)")
	if err != nil {log.Println("AddPekaMessageCache", err)}
}

func (d *DataBase) InsertPekaUserChannelsCache(data string, id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO peka_users_in_channels_cache"+id+" (requested) VALUES (?)", data)
	if err != nil {log.Println("InsertPekaUserChannelsCache", err)}
}

func (d *DataBase) AddGGMessageCache(id string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("INSERT INTO goodgame_message_cache"+id+" (id) VALUES (null)")
	if err != nil {log.Println("AddGGMessageCache", err)}
}

func (d *DataBase) GetGGMessagesCount(id string) int {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s := d.db.QueryRow("SELECT count(*) as cnt FROM goodgame_message_cache"+id)

	var cnt int
	s.Scan(&cnt)

	return cnt
}

func (d *DataBase) GetAllPekaUserInChannelsCache(id string) []string {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s, err := d.db.Query("SELECT requested FROM peka_users_in_channels_cache"+id)

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
	if err != nil {log.Println("InsertGGUserChannelsCache", err)}
}

func (d *DataBase) FlushGGUserChannelsCache() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	_, err := d.db.Exec("DELETE FROM goodgame_users_in_channels_cache")
	if err != nil {log.Println("FlushGGUserChannelsCache", err)}
}

func (d *DataBase) GetAllGoodGameUserInChannelsCache(id string) []string {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s, err := d.db.Query("SELECT requested FROM goodgame_users_in_channels_cache"+id)

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
	s := d.db.QueryRow("SELECT count(*) as cnt FROM peka_message_cache"+id)

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
	if err != nil {log.Println("InsertStatistics", err)}
}

func (d *DataBase) GetStatistics() []RowMap {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	s, _ := d.db.Query("SELECT externalTime, uniqUsers, messageCount, source FROM statistics WHERE externalTime != '' ORDER BY externalTime DESC")
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
