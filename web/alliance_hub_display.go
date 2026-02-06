// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the alliance hub display.

package web

import (
	"net/http"

	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
)

// Renders the alliance hub display showing the active status of the alliance's hub.
func (web *Web) allianceHubDisplayHandler(w http.ResponseWriter, r *http.Request) {
	if !web.enforceDisplayConfiguration(w, r, map[string]string{"alliance": "Red"}) {
		return
	}

	template, err := web.parseFiles("templates/alliance_hub_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "alliance_hub_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the alliance hub display client to receive status updates.
func (web *Web) allianceHubDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplay(r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client.
	ws.HandleNotifiers(
		display.Notifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier,
	)
}
