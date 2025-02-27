package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	c "github.com/ION-Smart/ion.mqtt/internal/controllers"
)

func main() {
	err := c.InitDB()
	if err != nil {
		log.Fatalln(err)
	}

	remontadores, err := c.ObtenerRemontadores(23, 0)
	if err != nil {
		log.Fatalln(err)
	}

	tiempo, err := c.ObtenerPersonasTransportadasHoyRemontador(remontadores[0].CodRemontador)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := json.Marshal(tiempo)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(bytes.NewBuffer(body))
}
