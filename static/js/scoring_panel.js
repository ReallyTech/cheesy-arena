// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;
let nearSide;
let committed = false;

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;
// True when teleop-only scoring controls should be available
let inTeleop = false;
// True when post-auto and in edit auto mode
let editingAuto = false;

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);
  const teams = (alliance === "red") ? [data.Match.Red1, data.Match.Red2, data.Match.Red3] : [data.Match.Blue1, data.Match.Blue2, data.Match.Blue3];
  for (let i = 0; i < 3; i++) {
    const pos = i + 1;
    $(`#tower-auto-${pos} .team-num`).text("Team " + (teams[i] || ""));
    $(`#tower-teleop-${pos} .team-num`).text("Team " + (teams[i] || ""));

    // Reset buttons on match load
    $(`#tower-auto-${pos}`).attr("data-selected", "false");
    $(`#tower-auto-${pos}-status`).text("None");
    $(`#tower-teleop-${pos}`).attr("data-selected", "false");
    $(`#tower-teleop-${pos}`).attr("data-level", "0");
    $(`#tower-teleop-${pos}-status`).text("None");
  }

  // Reset counters on match load
  $("#value-fuel-auto").text("0");
  $("#value-fuel-teleop").text("0");
  $("#hub-active-indicator").attr("data-hub-active", "false");
};

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
  translateMatchTime(data, function (state, stateText, countdown) {
    $("#matchState").text(stateText);
    $("#matchTime").text(getCountdownString(countdown));
  });
  switch (matchStates[data.MatchState]) {
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      inTeleop = false;
      committed = false;
      break;
    case "TELEOP_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      inTeleop = true;
      committed = false;
      break;
    case "POST_MATCH":
      if (!committed) {
        scoringAvailable = true;
        commitAvailable = true;
        inTeleop = true;
      }
      break;
    default:
      scoringAvailable = false;
      commitAvailable = false;
      inTeleop = false;
      committed = false;
  }
  updateUIMode();
};

// Refresh which UI controls are enabled/disabled
const updateUIMode = function () {
  $(".team-button").prop('disabled', false);
  $(".counter button").prop('disabled', false);
  $("#commit").prop('disabled', !commitAvailable);
  $(".container").attr("data-in-teleop", inTeleop && scoringAvailable);
}

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  let realtimeScore = (alliance === "red") ? data.Red : data.Blue;
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const pos = i + 1;
    const level = score.TowerLevels[i];
    const isAuto = score.TowerAuto[i];

    let levelText = "None";
    if (level === 1) levelText = "Level 1";
    if (level === 2) levelText = "Level 2";
    if (level === 3) levelText = "Level 3";

    // Auto button display
    $(`#tower-auto-${pos}-status`).text(isAuto ? "Auto" : "None");
    $(`#tower-auto-${pos}`).attr("data-selected", isAuto ? "true" : "false");

    // Teleop button display
    $(`#tower-teleop-${pos}-status`).text(levelText);
    $(`#tower-teleop-${pos}`).attr("data-selected", level > 0 ? "true" : "false");
    $(`#tower-teleop-${pos}`).attr("data-level", level.toString());
  }

  $(`#value-fuel-auto`).text(score.FuelAuto);
  $(`#value-fuel-teleop`).text(score.FuelTeleop);

  $("#hub-active-indicator").attr("data-hub-active", realtimeScore.HubActive);
};

const cycleTowerAuto = function (index) {
  // Toggle between Auto and None
  const isAuto = $(`#tower-auto-${index}-status`).text() !== "Auto";
  const level = parseInt($(`#tower-teleop-${index}`).attr("data-level"));

  websocket.send("tower", {
    TeamPosition: index,
    Level: level,
    IsAuto: isAuto
  });
}

const cycleTowerTeleop = function (index) {
  // Cycle between 0 -> 1 -> 2 -> 3 -> 0
  let level = parseInt($(`#tower-teleop-${index}`).attr("data-level"));
  const isAuto = $(`#tower-auto-${index}-status`).text() === "Auto";

  level = (level + 1) % 4;

  websocket.send("tower", {
    TeamPosition: index,
    Level: level,
    IsAuto: isAuto
  });
}

const incrementCounter = function (id) {
  let shift = id.replace("fuel-", "");
  websocket.send("fuel", { Shift: shift, Adjustment: 1 });
}

const decrementCounter = function (id) {
  let shift = id.replace("fuel-", "");
  websocket.send("fuel", { Shift: shift, Adjustment: -1 });
}

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function () {
  websocket.send("commitMatch");
  committed = true;
  updateUIMode();
};

$(function () {
  const pathParts = window.location.pathname.split("/");
  alliance = pathParts[pathParts.length - 1]; // "red" or "blue"
  $(".container").attr("data-alliance", alliance);

  // Attach event listeners
  $(document).on("click", ".tower-auto-btn", function () {
    cycleTowerAuto(parseInt($(this).data("index")));
  });
  $(document).on("click", ".tower-teleop-btn", function () {
    cycleTowerTeleop(parseInt($(this).data("index")));
  });
  $(document).on("click", ".counter .plus", function () {
    incrementCounter($(this).closest(".counter").data("id"));
  });
  $(document).on("click", ".counter .minus", function () {
    decrementCounter($(this).closest(".counter").data("id"));
  });
  $(document).on("click", "#commit", function () {
    commitMatchScore();
  });

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function (event) { handleMatchLoad(event.data); },
    matchTime: function (event) { handleMatchTime(event.data); },
    matchTiming: function (event) { handleMatchTiming(event.data); },
    realtimeScore: function (event) { handleRealtimeScore(event.data); },
    resetLocalState: function (event) { committed = false; updateUIMode(); },
  });
});

