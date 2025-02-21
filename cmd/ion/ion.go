package main

import (
	// "fmt"
	// "log"

	"github.com/ION-Smart/ion.mqtt/internal/config"
	ion "github.com/ION-Smart/ion.mqtt/internal/iondatabase"
)

func main() {
	conf := config.New()

	ion.GetConnection(conf)
}
