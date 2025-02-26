package main

import (
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

	segundos := c.ComprobarEnvioAlertaOcupacionRemontador(remontadores[0])

	fmt.Println(segundos)
}
