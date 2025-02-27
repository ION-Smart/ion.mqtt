package cvediahandlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ION-Smart/ion.mqtt/internal/controllers"
	cv "github.com/ION-Smart/ion.mqtt/pkg/cvevents"
)

func CrowdestCallback(msg string) error {
	fmt.Println("Crowd Estimation callback")
	msg = strings.TrimRight(strings.TrimLeft(msg, "["), "]")

	in := []byte(msg)
	var data cv.MessageCrowd

	err := json.Unmarshal(in, &data)
	if err != nil {
		return err
	}

	if data.InstanceId == "test crowd" {
		data.InstanceId = "0c527620-f5a5-45dc-4ffc-7696cf817fbe"
	}

	// fmt.Printf("Count: %d, Instance: %v\n", data.Count, data.InstanceId)
	controllers.InsertarOcupacionCrowdest(data)
	return nil
}

func SecurtCallback(msg string) error {
	fmt.Println("SecuRT callback")
	msg = strings.TrimRight(strings.TrimLeft(msg, "["), "]")

	in := []byte(msg)
	var data cv.MessageSecuRTest

	err := json.Unmarshal(in, &data)
	if err != nil {
		return err
	}

	event := data.Event
	fmt.Printf("Count: %d, Instance: %v\n", event.Extra.CurrentEntries, event.InstanceId)

	if data.Event.InstanceId == "secur" {
		data.Event.InstanceId = "bed1628a-63d5-e612-3808-78454a33a031"
	}
	controllers.InsertarOcupacionSecurt(data)

	return nil
}
