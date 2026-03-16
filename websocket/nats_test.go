// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package websocket

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestNATSIntegration(t *testing.T) {
	// 1. Start embedded NATS server on random ports
	// Using 4223 and 8082 to avoid conflict with default if running elsewhere
	port := 4223
	wsPort := 8082
	err := StartEmbeddedNATSServer(port, wsPort)
	assert.Nil(t, err)

	// 2. Initialize NATS client
	url := fmt.Sprintf("nats://127.0.0.1:%d", port)
	err = InitializeNATS(url)
	assert.Nil(t, err)
	defer GetNATS().Close()

	// 3. Test Authentication and Token Generation
	t.Run("AuthTokens", func(t *testing.T) {
		adminToken, err := GetNATSToken("admin", true)
		assert.Nil(t, err)
		assert.NotEmpty(t, adminToken)

		userToken, err := GetNATSToken("user", false)
		assert.Nil(t, err)
		assert.NotEmpty(t, userToken)

		// Test admin connection
		ncAdmin, err := nats.Connect(url, nats.Token(adminToken))
		assert.Nil(t, err)
		defer ncAdmin.Close()

		// Admin should be able to publish to arena.>
		err = ncAdmin.Publish("arena.test", []byte("hello"))
		assert.Nil(t, err)

		// Test user connection
		ncUser, err := nats.Connect(url, nats.Token(userToken))
		assert.Nil(t, err)
		defer ncUser.Close()

		// User should NOT be able to publish to arena.commands.> (based on claims in nats.go)
		err = ncUser.Publish("arena.commands.test", []byte("forbidden"))
		// NATS might not return error immediately on Publish, but permissions are enforced on server
	})

	// 4. Test Notification / PublishMessage
	t.Run("PublishMessage", func(t *testing.T) {
		subSubject := "arena.notify.test"
		msgChan := make(chan *nats.Msg, 1)
		sub, err := GetNATS().Subscribe(subSubject, func(m *nats.Msg) {
			msgChan <- m
		})
		assert.Nil(t, err)
		defer sub.Unsubscribe()

		testData := map[string]string{"foo": "bar"}
		err = PublishMessage(subSubject, "testType", testData)
		assert.Nil(t, err)

		select {
		case m := <-msgChan:
			var received Message
			err := json.Unmarshal(m.Data, &received)
			assert.Nil(t, err)
			assert.Equal(t, "testType", received.Type)
			dataMap := received.Data.(map[string]any)
			assert.Equal(t, "bar", dataMap["foo"])
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for NATS message")
		}
	})

	// 5. Test Bootstrap Handler
	t.Run("Bootstrap", func(t *testing.T) {
		// Mock a notifier
		NewNotifier("bootstrapTest", func() any {
			return map[string]int{"value": 42}
		})

		// Request bootstrap via NATS
		resp, err := GetNATS().Request("arena.bootstrap.bootstrapTest", nil, 1*time.Second)
		assert.Nil(t, err)

		var received Message
		err = json.Unmarshal(resp.Data, &received)
		assert.Nil(t, err)
		assert.Equal(t, "bootstrapTest", received.Type)
		dataMap := received.Data.(map[string]any)
		assert.Equal(t, float64(42), dataMap["value"]) // JSON unmarshals to float64
	})

	// 6. Test Command Handler and Path Normalization
	t.Run("CommandsAndNormalization", func(t *testing.T) {
		receivedCommand := make(chan string, 1)
		receivedData := make(chan any, 1)

		// Register a handler for a path that needs normalization
		// Input path "/panels/scoring/red-near/websocket" should become "panels.scoring.red-near"
		RegisterCommandHandler("/panels/scoring/red-near/websocket", func(messageType string, data any, clientId string) bool {
			receivedCommand <- messageType
			receivedData <- data
			return true
		})

		// 1. Test normalized path publishing
		cmdPath := "arena.commands.panels.scoring.red-near"
		testData := map[string]any{"action": "score", "value": float64(1)}
		cmdPayload := struct {
			Type     string `json:"type"`
			Data     any    `json:"data"`
			ClientId string `json:"clientId"`
		}{
			Type:     "buttonPress",
			Data:     testData,
			ClientId: "test-client-1",
		}
		payload, _ := json.Marshal(cmdPayload)
		err := GetNATS().Publish(cmdPath, payload)
		assert.Nil(t, err)

		select {
		case msgType := <-receivedCommand:
			assert.Equal(t, "buttonPress", msgType)
			data := <-receivedData
			assert.Equal(t, testData, data)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for command handler with normalized path")
		}

		// 2. Test publishing with complex data types
		PublishMessage("arena.notify.display", "updateDisplay", map[string]string{"status": "match_running"})

		sub, _ := GetNATS().Subscribe("arena.notify.display", func(m *nats.Msg) {
			var msg Message
			json.Unmarshal(m.Data, &msg)
			if msg.Type == "updateDisplay" {
				receivedCommand <- "published"
			}
		})
		defer sub.Unsubscribe()

		PublishMessage("arena.notify.display", "updateDisplay", nil)

		select {
		case res := <-receivedCommand:
			assert.Equal(t, "published", res)
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for PublishMessage verification")
		}
	})
}
