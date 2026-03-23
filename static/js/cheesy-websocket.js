// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Shared code for initiating websocket connections back to the server for full-duplex communication.

var DisplayIdentity = (function () {
  var storageKey = "displayId";

  var generateDisplayId = function () {
    if (window.crypto !== undefined && typeof window.crypto.randomUUID === "function") {
      return window.crypto.randomUUID();
    }

    // Fallback for browsers without crypto.randomUUID().
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (character) {
      var randomValue = Math.floor(Math.random() * 16);
      var value = character === "x" ? randomValue : (randomValue & 0x3) | 0x8;
      return value.toString(16);
    });
  };

  var readStoredDisplayId = function () {
    try {
      return window.sessionStorage.getItem(storageKey);
    } catch (error) {
      return null;
    }
  };

  var storeDisplayId = function (displayId) {
    try {
      window.sessionStorage.setItem(storageKey, displayId);
    } catch (error) {
      // Ignore storage errors; the websocket can still use the in-memory ID.
    }
  };

  var syncUrl = function (displayId) {
    var urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get(storageKey) === displayId) {
      return;
    }

    urlParams.set(storageKey, displayId);
    var newSearch = urlParams.toString();
    var newUrl = window.location.pathname + (newSearch !== "" ? "?" + newSearch : "") + window.location.hash;
    if (window.history !== undefined && typeof window.history.replaceState === "function") {
      window.history.replaceState(null, "", newUrl);
    }
  };

  var getDisplayId = function () {
    var urlDisplayId = new URLSearchParams(window.location.search).get(storageKey);
    var displayId = urlDisplayId || readStoredDisplayId() || generateDisplayId();
    storeDisplayId(displayId);
    if (urlDisplayId !== displayId) {
      syncUrl(displayId);
    }
    return displayId;
  };

  return {
    getDisplayId: getDisplayId,
    storeDisplayId: storeDisplayId,
  };
})();

var CheesyWebsocket = function (path, events) {
  var that = this;
  var protocol = "ws://";
  if (window.location.protocol === "https:") {
    protocol = "wss://";
  }
  var url = protocol + window.location.hostname;
  if (window.location.port !== "") {
    url += ":" + window.location.port;
  }
  url += path;

  // Append the page's query string to the websocket URL.
  url += window.location.search;

  // Insert a default error-handling event if a custom one doesn't already exist.
  if (!events.hasOwnProperty("error")) {
    events.error = function (event) {
      // Data is just an error string.
      console.log(event.data);
      alert(event.data);
    };
  }

  // Parse the display parameters that will be present in the query string if this is a display.
  var displayId = DisplayIdentity.getDisplayId();

  // Insert an event to allow the server to force-reload the client for any display.
  events.reload = function (event) {
    if (event.data === null || event.data === displayId) {
      location.reload();
    }
  };

  // Insert an event to allow reconfiguration if this is a display.
  if (!events.hasOwnProperty("displayConfiguration")) {
    events.displayConfiguration = function (event) {
      var newUrl = event.data;

      // Reload the display if the configuration has changed.
      if (newUrl !== window.location.pathname + window.location.search) {
        window.location = newUrl;
      }
    };
  }

  this.connect = function () {
    this.websocket = $.websocket(url, {
      open: function () {
        console.log("Websocket connected to the server at " + url + ".")
      },
      close: function () {
        console.log("Websocket lost connection to the server. Reconnecting in 3 seconds...");
        setTimeout(that.connect, 3000);
      },
      events: events
    });
  };

  this.send = function (type, data) {
    this.websocket.send(type, data);
  };

  this.connect();
};
