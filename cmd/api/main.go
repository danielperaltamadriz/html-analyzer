package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/danielperaltamadriz/home24/api"
)

func main() {
	server := api.NewAPI(api.APIConfig{})
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Println("Starting server...")
		err := server.Start()
		if err != nil {
			log.Fatal("Failed to start server: ", err)
		}
		wg.Done()
	}()

	<-ctx.Done()

	err := server.Shutdown()
	if err != nil {
		log.Fatal("Failed to stop server: ", err)
	}
	wg.Wait()
	fmt.Println("Server stopped")
}
