package model

import (
	"constant"
	"math/rand"
	"model/db"
	"time"
	"util/log"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

var (
	DefaultSelector bson.M = bson.M{}
	logger                 = log.GetLogger()
)

func getRedisDefaultExpire() int64 {
	rand.Seed(time.Now().UnixNano())
	return constant.RedisDefaultExpire + rand.Int63n(constant.RedisDefaultRandExpire)
}

/****************************************** log ****************************************/

func writeIndexLog(funcName, errMsg string, err error) {
	writeLog("index.go", funcName, errMsg, err)
}

func writeLog(fileName, funcName, errMsg string, err error) {
	logger.WithFields(logrus.Fields{
		"package":  "model",
		"file":     fileName,
		"function": funcName,
		"err":      err,
	}).Warn(errMsg)
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
