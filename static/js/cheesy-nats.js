// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Modified version of CheesyWebsocket that uses NATS over WebSockets.
// Falls back to regular WebSocket if NATS is unavailable.

var CheesyNats = function (path, events, natsUrl) {
  var that = this;
  this.events = events;
  this.nc = null;
  this.sc = null; // String codec

  // Default NATS URL if not provided
  natsUrl = natsUrl || "/api/nats/token";

  // Insert a default error-handling event if a custom one doesn't already exist.
  if (!events.hasOwnProperty("error")) {
    events.error = function (msg) {
      console.error("NATS Error:", msg);
    };
  }

  // Parse the display parameters.
  var displayId = new URLSearchParams(window.location.search).get("displayId");

  // Reload event
  if (!events.hasOwnProperty("reload")) {
      events.reload = function (data) {
        if (data === null || data === displayId) {
          location.reload();
        }
      };
  }

  this.connect = async function () {
    try {
      if (typeof nats === 'undefined') {
        console.warn("nats.js not loaded. Falling back to legacy WebSocket.");
        new CheesyWebsocket(path, events);
        return;
      }

      // Fetch NATS token and URL from backend
      const response = await fetch(natsUrl);
      if (!response.ok) {
        throw new Error("Failed to fetch NATS token from " + natsUrl);
      }
      const auth = await response.json();

      this.sc = nats.StringCodec();
      this.nc = await nats.connect({
        servers: auth.url || "ws://" + window.location.hostname + ":8081",
        token: auth.token,
        maxReconnectAttempts: -1,
      });
      console.log("NATS connected with token to " + (auth.url || "ws://" + window.location.hostname + ":8081"));

      this.nc.closed().then((err) => {
        console.log("NATS connection closed. Reconnecting in 3 seconds...", err ? err : "");
        setTimeout(() => that.connect(), 3000);
      });

      // Subscribe to all updates
      const sub = this.nc.subscribe("cheesy.updates.>");
      (async () => {
        for await (const m of sub) {
          const payload = JSON.parse(this.sc.decode(m.data));
          if (this.events[payload.type]) {
            this.events[payload.type]({ data: payload.data });
          }
        }
      })();

      // Bootstrap current state for each registered event
      for (const messageType in this.events) {
        if (messageType === 'open' || messageType === 'close' || messageType === 'error') continue;
        
        try {
          const msg = await this.nc.request("cheesy.bootstrap." + messageType, nats.Empty, { timeout: 1000 });
          const payload = JSON.parse(this.sc.decode(msg.data));
          this.events[payload.type]({ data: payload.data });
        } catch (err) {
          // It's fine if some don't support bootstrap
        }
      }

      if (this.events.open) this.events.open();

    } catch (err) {
      console.error("Failed to connect to NATS:", err);
      console.log("Falling back to legacy WebSocket.");
      new CheesyWebsocket(path, events);
    }
  };

  this.send = function (type, data) {
    if (this.nc) {
      this.nc.publish("cheesy.commands." + type, this.sc.encode(JSON.stringify({ type: type, data: data })));
    }
  };

  this.connect();
};
