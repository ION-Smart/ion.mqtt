package main

import (
	"encoding/json"
	"fmt"
	"log"

	c "github.com/ION-Smart/ion.mqtt/internal/controllers"
	// "github.com/ION-Smart/ion.mqtt/internal/projectpath"
)

func main() {
	err := c.InitDB()
	if err != nil {
		log.Fatalln(err)
	}

	alertas, err := c.ObtenerAlertasRemontadoresSkiParam(923, 0, 10)
	if err != nil {
		log.Fatalln(err)
	}

	body := c.BodyAlertaSkiSocket{
		Server:  "localhost",
		Alertas: alertas,
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}

	var al2 c.BodyAlertaSkiSocket
	err = json.Unmarshal(data, &al2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(al2)
}
