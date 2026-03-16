// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// NATS integration for Cheesy Arena.

package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var (
	nc         *nats.Conn
	natsMutex  sync.Mutex
	natsServer string
	ns         *server.Server
)

// StartEmbeddedNATSServer starts an in-process NATS server with WebSocket support.
func StartEmbeddedNATSServer(port int, wsPort int) error {
	opts := &server.Options{
		Host: "127.0.0.1",
		Port: port,
		Websocket: server.WebsocketOpts{
			Host:  "0.0.0.0",
			Port:  wsPort,
			NoTLS: true,
		},
	}

	var err error
	ns, err = server.NewServer(opts)
	if err != nil {
		return err
	}

	// Go routine to start the server
	go ns.Start()

	if !ns.ReadyForConnections(10 * time.Second) {
		return nats.ErrNoServers
	}

	log.Printf("Embedded NATS Server started (TCP: 127.0.0.1:%d, WS: 0.0.0.0:%d)", port, wsPort)
	return nil
}

// InitializeNATS sets up the global NATS connection.
func InitializeNATS(url string) error {
	natsMutex.Lock()
	defer natsMutex.Unlock()

	var err error
	nc, err = nats.Connect(url,
		nats.Name("Cheesy Arena"),
		nats.Timeout(10*time.Second),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return err
	}
	natsServer = url
	log.Printf("NATS connected to %s", url)

	// Setup Respond handler for bootstrap requests
	setupBootstrapHandler()

	return nil
}

// GetNATS returns the global NATS connection.
func GetNATS() *nats.Conn {
	return nc
}

// PublishMessage sends a message to a NATS subject.
func PublishMessage(subject string, messageType string, data any) error {
	if nc == nil {
		return nil // NATS not initialized
	}

	msg := Message{
		Type: messageType,
		Data: data,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return nc.Publish(subject, payload)
}

func setupBootstrapHandler() {
	// Listen for requests on "cheesy.bootstrap.>"
	// Clients can request initial state for a specific notifier
	nc.Subscribe("cheesy.bootstrap.>", func(m *nats.Msg) {
		// Subject will be cheesy.bootstrap.<messageType>
		parts := nats.TokenizeSubject(m.Subject)
		if len(parts) < 3 {
			return
		}
		messageType := parts[2]

		// Find the notifier and get its state
		state := GetNotifierState(messageType)
		if state != nil {
			msg := Message{
				Type: messageType,
				Data: state,
			}
			payload, _ := json.Marshal(msg)
			m.Respond(payload)
		}
	})
}
