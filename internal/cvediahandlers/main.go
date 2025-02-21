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

	fmt.Printf("Count: %d, Instance: %v\n", data.Count, data.InstanceId)

	controllers.InsertarOcupacionCrowdest(data)
	return nil
}

func SecurtCallback(msg string) error {
	fmt.Println("SecuRT callback")
	msg = strings.TrimRight(strings.TrimLeft(msg, "["), "]")

	in := []byte(msg)
	var data cv.MessageSecuRT

	err := json.Unmarshal(in, &data)
	if err != nil {
		return err
	}

	event := data.Events[0]
	fmt.Printf("Count: %d, Instance: %v\n", event.Extra.CurrentEntries, event.InstanceId)

	controllers.InsertarOcupacionSecurt(data)

	return nil
}
