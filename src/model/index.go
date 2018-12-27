package model

import (
	"constant"
	"math/rand"
	"model/db"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var (
	DefaultSelector bson.M = bson.M{}
)

func getRedisDefaultExpire() int64 {
	rand.Seed(time.Now().UnixNano())
	return constant.RedisDefaultExpire + rand.Int63n(constant.RedisDefaultRandExpire)
}

/****************************************** db basic action ****************************************/

func updateDoc(tableName string, query, update interface{}) error {
	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(tableName)
	return table.Update(query, update)
}

func updateDocs(tableName string, query, update interface{}) (interface{}, error) {
	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(tableName)
	return table.UpdateAll(query, update)
}

func insertDocs(tableName string, docs ...interface{}) error {
	cntrl := db.NewCloneMgoDBCntlr()
	defer cntrl.Close()
	table := cntrl.GetTable(tableName)
	return table.Insert(docs...)
}
