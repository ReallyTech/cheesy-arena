// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

// Go version 1.22 or newer is required.
//go:build go1.22

package main

import (
	"fmt"
	"log"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/web"
	"github.com/Team254/cheesy-arena/websocket"
)

const eventDbPath = "./event.db"
const httpPort = 8080
const natsPort = 4222
const natsWsPort = 8081 // Port for NATS-over-WebSockets

// Main entry point for the application.
func main() {
	// Start Embedded NATS Server
	err := websocket.StartEmbeddedNATSServer(natsPort, natsWsPort)
	if err != nil {
		log.Printf("Warning: Failed to start embedded NATS server (%v).", err)
	}

	// Initialize NATS Client (connect to self)
	err = websocket.InitializeNATS(fmt.Sprintf("nats://127.0.0.1:%d", natsPort))
	if err != nil {
		log.Printf("Warning: Failed to connect to NATS (%v). WebSocket fallback will be used for distribution.", err)
	}

	arena, err := field.NewArena(eventDbPath)
	if err != nil {
		log.Fatalln("Error during startup: ", err)
	}

	// Start the web server in a separate goroutine.
	web := web.NewWeb(arena)
	go web.ServeWebInterface(httpPort)

	// Run the arena state machine in the main thread.
	arena.Run()
}
