// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Modified version of CheesyWebsocket that uses NATS over WebSockets.
// Falls back to regular WebSocket if NATS is unavailable.

var CheesyNats = function (path, events, options) {
  var that = this;
  this.events = events;
  this.nc = null;
  this.sc = null; // String codec

  options = options || {};
  // Default NATS URL if not provided
  var natsUrl = options.natsUrl || "/api/nats/token";
  
  // Construct a unique command subject based on the URL path
  // E.g., /panels/scoring/red_near/websocket -> arena.commands.panels.scoring.red_near
  var pathParts = path.split('/').filter(function(p) { return p && p !== 'websocket'; });
  this.commandSubject = "arena.commands." + pathParts.join('.');

  // Insert a default error-handling event if a custom one doesn't already exist.
  if (!events.hasOwnProperty("error")) {
    events.error = function (msg) {
      console.error("NATS Error:", msg);
    };
  }

  // Get or generate a persistent client ID for this session
  var clientId = sessionStorage.getItem("nats_client_id");
  if (!clientId) {
    clientId = crypto.randomUUID();
    sessionStorage.setItem("nats_client_id", clientId);
  }

  // Reload event
  if (!events.hasOwnProperty("reload")) {
      events.reload = function (data) {
        if (data === null || data === clientId) {
          location.reload();
        }
      };
  }

  this.connect = async function () {
    try {
      if (typeof nats === 'undefined') {
        console.warn("nats.js not loaded. Cannot connect.");
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

      // Send initial registration or connection status if needed via WebSocket path
      // but for NATS we are already connected. 
      // We can trigger the server to recognize this clientId by sending a registration command.
      var subscriptions = Object.keys(this.events).filter(function(k) {
        return k !== 'open' && k !== 'close' && k !== 'error';
      });
      this.send("register", { clientId: clientId, path: path, subscriptions: subscriptions });

      this.nc.closed().then((err) => {
        console.log("NATS connection closed. Reconnecting in 3 seconds...", err ? err : "");
        setTimeout(() => that.connect(), 3000);
      });

      // Subscribe to client-specific updates
      const clientSub = this.nc.subscribe("arena.clients." + clientId + ".>");
      (async () => {
        for await (const m of clientSub) {
          const payload = JSON.parse(this.sc.decode(m.data));
          if (this.events[payload.type]) {
            this.events[payload.type](payload.data);
          }
        }
      })();

      // Bootstrap current state for each registered event
      for (const messageType in this.events) {
        if (messageType === 'open' || messageType === 'close' || messageType === 'error') continue;

        try {
          const msg = await this.nc.request("arena.bootstrap." + messageType, nats.Empty, { timeout: 1000 });
          const payload = JSON.parse(this.sc.decode(msg.data));
          this.events[payload.type](payload.data);
        } catch (err) {
          // It's fine if some don't support bootstrap
        }
      }

      if (this.events.open) this.events.open();

    } catch (err) {
      console.error("Failed to connect to NATS:", err);
    }
  };

  this.send = function (type, data) {
    if (this.nc) {
      this.nc.publish(this.commandSubject, this.sc.encode(JSON.stringify({ type: type, data: data, clientId: clientId })));
    }
  };

  this.connect();
};
