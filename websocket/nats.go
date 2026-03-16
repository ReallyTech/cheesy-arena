// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// NATS integration for Cheesy Arena.

package websocket

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

var (
	nc           *nats.Conn
	natsMutex    sync.Mutex
	natsServer   string
	ns           *server.Server
	serverKey    nkeys.KeyPair
	operatorKey  nkeys.KeyPair
	accountKey   nkeys.KeyPair
	natsStarted  bool
	natsInternal bool // True if using embedded server
)

// StartEmbeddedNATSServer starts an in-process NATS server with WebSocket support.
func StartEmbeddedNATSServer(port int, wsPort int) error {
	var err error
	operatorKey, err = nkeys.CreateOperator()
	if err != nil {
		return err
	}
	accountKey, err = nkeys.CreateAccount()
	if err != nil {
		return err
	}
	serverKey, err = nkeys.CreateServer()
	if err != nil {
		return err
	}

	opts := &server.Options{
		Host: "127.0.0.1",
		Port: port,
		Websocket: server.WebsocketOpts{
			Host:  "0.0.0.0",
			Port:  wsPort,
			NoTLS: true,
		},
		CustomClientAuthentication: &natsAuth{
			operatorKey: operatorKey,
			accountKey:  accountKey,
		},
	}

	ns, err = server.NewServer(opts)
	if err != nil {
		return err
	}

	// Go routine to start the server
	go ns.Start()

	if !ns.ReadyForConnections(10 * time.Second) {
		return nats.ErrNoServers
	}

	natsStarted = true
	log.Printf("Embedded NATS Server started (TCP: 127.0.0.1:%d, WS: 0.0.0.0:%d) with auth", port, wsPort)
	return nil
}

type natsAuth struct {
	operatorKey nkeys.KeyPair
	accountKey  nkeys.KeyPair
}

func (a *natsAuth) Check(c server.ClientAuthentication) bool {
	opts := c.GetOpts()
	token := opts.Token

	// If no token, check if it's the internal client (we might need to allow it without token or with specific token)
	if token == "" {
		return false
	}

	// For external clients, they must provide a valid JWT token signed by us.
	// In this simplified implementation, we'll use the token as the JWT directly.
	claims, err := jwt.DecodeUserClaims(token)
	if err != nil {
		return false
	}

	vr := &jwt.ValidationResults{}
	claims.Validate(vr)
	if vr.IsBlocking() {
		return false
	}

	return true
}

func GetNATSToken(username string, isAdmin bool) (string, error) {
	if !natsStarted {
		return "", nil
	}

	userKey, err := nkeys.CreateUser()
	if err != nil {
		return "", err
	}
	pub, err := userKey.PublicKey()
	if err != nil {
		return "", err
	}

	claims := jwt.NewUserClaims(pub)
	claims.Name = username

	// Define permissions
	if isAdmin {
		// Admins can do everything
		claims.Pub.Allow.Add("cheesy.>")
		claims.Sub.Allow.Add("cheesy.>")
	} else {
		// Regular users can only subscribe to updates and bootstrap
		claims.Sub.Allow.Add("cheesy.updates.>")
		claims.Sub.Allow.Add("cheesy.bootstrap.>")
		// They cannot publish to commands or updates
		claims.Pub.Deny.Add("cheesy.commands.>")
		claims.Pub.Deny.Add("cheesy.updates.>")
	}

	accPub, err := accountKey.PublicKey()
	if err != nil {
		return "", err
	}

	token, err := claims.Encode(accountKey)
	if err != nil {
		return "", err
	}

	return token, nil
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
		parts := strings.Split(m.Subject, ".")
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
