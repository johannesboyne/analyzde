package main

import (
	"log"
	"time"
)

func SaveToMongoDB(mngoC MngoConnection, event EventStruct) bool {
	start := time.Now()
	c := mngoC.C(event.Id)
	err := c.Insert(event)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("saved to MongoDB, time: %s", time.Since(start))
	return true
}
func PrintToStdout(event EventStruct) bool {
	log.Println(event)
	return true
}
