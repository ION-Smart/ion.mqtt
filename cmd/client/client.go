package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	ionmqtt "github.com/ION-Smart/ion.mqtt/pkg/ionmqtt"
)

func main() {
	broker := flag.String("broker", "tcp://test.mosquitto.org:1883", "Broker URI. ex: tcp://10.10.1.1:1883")
	user := flag.String("user", "", "Broker username for authentication")
	password := flag.String("password", "", "Broker password for authentication")
	id := "clientid"
	store := ":memory:" // The Store Directory (default use memory store)

	cleansess := false
	var qos int = 0 // The Quality of Service 0,1,2 (default 0)

	choke := make(chan [2]string)

	flag.Parse()

	client, err := ionmqtt.ConnectToBroker(*broker, *user, *password, id, store, cleansess, choke)

	if err != nil {
		log.Fatal(err)
		return
	}

	go ionmqtt.ListenTopic(client, "crowdest", qos, choke)
	go ionmqtt.ListenTopic(client, "securt", qos, choke)

	go forever()

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	//time for cleanup before exit
	fmt.Println("Adios!")
}

func forever() {
	for {
		time.Sleep(time.Second)
	}
}
