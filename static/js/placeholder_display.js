// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the placeholder display.

var websocket;

$(function () {
  // Read the configuration for this display from the session storage (clientId).
  var clientId = sessionStorage.getItem("nats_client_id");
  $("#displayId").text(clientId);
  var urlParams = new URLSearchParams(window.location.search);
  var nickname = urlParams.get("nickname");
  if (nickname !== null) {
    $("#displayNickname").text(nickname);
  }

  // Set up the websocket back to the server.
  websocket = new CheesyNats("/display/websocket", {});
});
