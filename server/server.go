package main

import (
	"log"
)

func server() {
	broker := "tcp://127.0.0.1:1883"
	user := ""
	password := ""
	id := "testgoid"
	store := ":memory:" // The Store Directory (default use memory store)

	cleansess := false
	var qos int = 0 // The Quality of Service 0,1,2 (default 0)

	choke := make(chan [2]string)

	client, err := connectToBroker(broker, user, password, id, store, cleansess, choke)

	if err != nil {
		log.Fatal(err)
		return
	}

	listenTopic(client, "crowdest", qos, choke)
	listenTopic(client, "securt", qos, choke)
}
