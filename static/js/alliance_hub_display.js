// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the alliance hub display.

var alliance = "";

$(document).ready(function () {
    var params = new URLSearchParams(window.location.search);
    alliance = params.get("alliance") || "Red";
    // Normalize alliance to "Red" or "Blue"
    alliance = alliance.charAt(0).toUpperCase() + alliance.slice(1).toLowerCase();
    $("body").attr("data-alliance", alliance.toLowerCase());

    var events = {
        matchTime: function (event) {
            translateMatchTime(event.data, function (matchState, matchStateText, countdownSec) {
                $("#match-state").text(matchStateText);
                $("#match-time").text(getCountdownString(countdownSec));
            });
        },
        matchTiming: function (event) {
            handleMatchTiming(event.data);
        },
        realtimeScore: function (event) {
            var data = event.data;
            var allianceData = data[alliance];
            if (allianceData) {
                var hubActive = allianceData.HubActive;
                $("#hub-status-text").text(hubActive ? "ACTIVE" : "INACTIVE");
                $("#hub-container").attr("data-hub-active", hubActive);

                if (allianceData.ScoreSummary) {
                    $("#fuel-count").text(allianceData.ScoreSummary.TotalFuel);
                    $("#fuel-threshold").text(allianceData.ScoreSummary.FuelNextRPThreshold);
                }
            }
        }
    };

    new CheesyWebsocket("/displays/alliance_hub/websocket", events);
});
