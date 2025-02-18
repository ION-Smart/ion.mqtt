package main

import (
	"fmt"
	"log"
	"time"

	ionmqtt "github.com/ION-Smart/ion.mqtt/pkg/ionmqtt"
)

func main() {
	broker := "tcp://127.0.0.1:1883"
	user := ""
	password := ""
	id := "serverid"
	store := ":memory:" // The Store Directory (default use memory store)
	var qos int = 0     // The Quality of Service 0,1,2 (default 0)
	cleansess := false

	choke := make(chan [2]string)

	client, err := ionmqtt.ConnectToBroker(broker, user, password, id, store, cleansess, choke)

	if err != nil {
		log.Fatal(err)
		return
	}

	keepGoing := true
	for keepGoing {
		var topic, payload string
		fmt.Print("Escribe el topic y el payload a enviar: ")
		_, err := fmt.Scanln(&topic, &payload)

		if err != nil {
			log.Fatal(err)
			return
		}

		ionmqtt.PublishTopic(client, topic, payload, qos)

		var follow byte
		follow = 1

		fmt.Print("Quieres seguir enviando eventos? (1/x) ")
		_, err = fmt.Scanln(&follow)

		if err != nil && follow != 1 {
			log.Fatal(err)
			keepGoing = false
		}
	}

	fmt.Println("Adios!")
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}
