package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func GetTotalById(mngoC MngoConnection, w http.ResponseWriter, id string, daysBack time.Duration) {
	c := mngoC.C(id)
	job := &mgo.MapReduce{
		Map:    "function () { emit(this.path, 1); }",
		Reduce: "function (path, count) { return Array.sum(count); }",
	}
	var result []struct {
		Path  string "_id"
		Value int
	}
	dateString := time.Now().Add(-1 * daysBack)
	log.Println(dateString)
	c.Find(bson.M{"timestamp": bson.M{"$gt": dateString}}).MapReduce(job, &result)
	binjson, _ := json.Marshal(result)
	w.Write(binjson)
}

func GetSeriesById(mngoC MngoConnection, w http.ResponseWriter, id string, daysBack time.Duration) {
	c := mngoC.C(id)
	dateString := time.Now().Add(-1 * daysBack)
	var result []struct {
		Id        string    "id"
		Path      string    "path"
		Timestamp time.Time "timestamp"
	}
	c.Find(bson.M{"timestamp": bson.M{"$gt": dateString}}).Iter().All(&result)
	binjson, _ := json.Marshal(result)
	w.Write(binjson)
}
