package main

import (
	"log"

	"github.com/doubleunion/accesscontrol/router"
	rpio "github.com/stianeikeland/go-rpio/v4"
)

func main() {
	err := rpio.Open()
	if err != nil {
		log.Fatalf("failed to initialize gpio: %+v", err)
	}
	defer rpio.Close()

	router.RunRouter()
}
