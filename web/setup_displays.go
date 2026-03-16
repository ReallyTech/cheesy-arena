// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for configuring the field displays.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

// Shows the displays configuration page.
func (web *Web) displaysGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_displays.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		DisplayTypeNames map[field.DisplayType]string
	}{web.arena.EventSettings, field.DisplayTypeNames}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the display configuration page to send control commands and receive status updates.
func (web *Web) displaysWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.DisplayConfigurationNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		err = web.handleDisplaysCommand(messageType, data)
		if err != nil {
			ws.WriteError(err.Error())
		}
	}
}

func (web *Web) handleDisplaysCommand(messageType string, data interface{}) error {
	switch messageType {
	case "configureDisplay":
		var displayConfig field.DisplayConfiguration
		err := mapstructure.Decode(data, &displayConfig)
		if err != nil {
			return err
		}
		if err = web.arena.UpdateDisplay(displayConfig); err != nil {
			return err
		}
	case "reloadDisplay":
		clientId, ok := data.(string)
		if !ok {
			return fmt.Errorf("Failed to parse '%s' message.", messageType)
		}
		// Notify by clientId channel
		websocket.PublishMessage("arena.clients."+clientId+".reload", "reload", clientId)
	case "reloadAllDisplays":
		web.arena.ReloadDisplaysNotifier.Notify()
		// Also notify all clients via wildcard if needed?
		// For now, Notify() already sends to arena.notify.reload
	default:
		return fmt.Errorf("Invalid message type '%s'.", messageType)
	}
	return nil
}
