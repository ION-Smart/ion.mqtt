package cvediahandlers

import (
	"encoding/json"
	"fmt"
	"strings"

	cv "github.com/ION-Smart/ion.mqtt/pkg/cvevents"
)

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

	return nil
}
