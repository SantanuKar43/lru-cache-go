package main

import (
	"fmt"
	"log"
	"net"
	"github.com/SantanuKar43/lru-cache-go/cache"
	"github.com/SantanuKar43/lru-cache-go/config"
	"github.com/SantanuKar43/lru-cache-go/connection"
)

func main() {
	config := config.GetConfig()
	cache := cache.Init(config.Capacity, config.WalSize, config.FlushStrategy)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatal("Error occurred while starting server", err)
	}

	log.Printf("started server on :%d", config.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Print("Unable to accept connection", err)
			continue
		}
		log.Printf("received connection: %s", conn.RemoteAddr())
		go connection.HandleConnection(conn, cache)
	}
}

