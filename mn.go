package main

import (
	"os"

	"gopkg.in/mgo.v2"
)

type MngoConnection struct {
	Session     *mgo.Session
	Opensession bool
}

func (connection *MngoConnection) open() bool {
	var err error
	if connection.Opensession == false {
		connection.Session, err = mgo.Dial(os.Getenv("MONGODB_URL"))
		if err != nil {
			panic(err)
		}
	}
	return true
}

func (connection *MngoConnection) C(id string) *mgo.Collection {
	if connection.Opensession == false {
		connection.open()
	}
	return connection.Session.DB("analyzde").C(id)
}
