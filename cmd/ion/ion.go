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

	rest, err := c.ObtenerRestaurantes(0, 0)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(rest)
}
