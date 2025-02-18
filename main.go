package ion_mqtt

import (
	"fmt"
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func logs() {
	MQTT.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	MQTT.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	MQTT.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	MQTT.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
}

func connectToBroker(broker, user, password, id, store string, cleansess bool, choke chan [2]string) (MQTT.Client, error) {
	fmt.Printf("Connection Info:\n")
	fmt.Printf("\tbroker:    %s\n", broker)
	fmt.Printf("\tclientid:  %s\n", id)
	fmt.Printf("\tuser:      %s\n", user)
	fmt.Printf("\tpassword:  %s\n", password)
	fmt.Printf("\tcleansess: %v\n", cleansess)
	fmt.Printf("\tstore:     %s\n", store)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(id)
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.SetCleanSession(cleansess)
	if store != ":memory:" {
		opts.SetStore(MQTT.NewFileStore(store))
	}

	fmt.Println("Trying to connect to MQTT Server")

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()

		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	return client, nil
}

func listenTopic(client MQTT.Client, topic string, qos int, choke chan [2]string) {
	// Connect, Subscribe, Publish etc..
	if topic == "" {
		fmt.Println("Invalid setting for -topic, must not be empty")
		return
	}

	fmt.Printf("\ttopic:     %s\n", topic)

	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	for {
		incoming := <-choke
		fmt.Printf("Received topic: %s message: %s\n", incoming[0], incoming[1])
	}
}

func publishTopic(client MQTT.Client, topic string, qos int, choke chan [2]string) {
	// Connect, Subscribe, Publish etc..
	if topic == "" {
		fmt.Println("Invalid setting for -topic, must not be empty")
		return
	}

	fmt.Printf("\ttopic:     %s\n", topic)

	if token := client.Subscribe(topic, byte(qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	for {
		incoming := <-choke
		fmt.Printf("Received topic: %s message: %s\n", incoming[0], incoming[1])
	}
}
