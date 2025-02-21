package cvediahandlers

import (
	"encoding/json"
	"fmt"
	"strings"

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
	return nil
}
