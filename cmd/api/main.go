package main

import (
	"log"

	"github.com/danielperaltamadriz/home24/internal"
)

func main() {
	server := internal.NewAPI(internal.APIConfig{})
	err := server.Start()
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
