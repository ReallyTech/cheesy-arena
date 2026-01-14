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

let localFoulCounts = {
  "red-minor": 0,
  "blue-minor": 0,
  "red-major": 0,
  "blue-major": 0,
}

// Handle controls to open/close the endgame dialog
const endgameDialog = $("#endgame-dialog")[0];
const showEndgameDialog = function () {
  endgameDialog.showModal();
}
const closeEndgameDialog = function () {
  endgameDialog.close();
}
const closeEndgameDialogIfOutside = function (event) {
  if (event.target === endgameDialog) {
    closeEndgameDialog();
  }
}

const foulsDialog = $("#fouls-dialog")[0];
const showFoulsDialog = function () {
  foulsDialog.showModal();
}
const closeFoulsDialog = function () {
  foulsDialog.close();
}
$(document).on("click", "#fouls-dialog", function (event) {
  if (event.target === foulsDialog) {
    closeFoulsDialog();
  }
});

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);
  const teams = (alliance === "red") ? [data.Match.Red1, data.Match.Red2, data.Match.Red3] : [data.Match.Blue1, data.Match.Blue2, data.Match.Blue3];
  for (let i = 0; i < 3; i++) {
    const pos = i + 1;
    $(`#tower-auto-${pos} .team-num`).text("Team " + (teams[i] || ""));
    $(`#tower-teleop-${pos} .team-num`).text("Team " + (teams[i] || ""));
  }
};

const addFoul = function (alliance, type) {
  const isMajor = type === "tech";
  websocket.send("addFoul", {Alliance: alliance, IsMajor: isMajor});
}

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
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
  $(".team-button").prop('disabled', !scoringAvailable);
  $("#tower-teleop .team-button").prop('disabled', !inTeleop || !scoringAvailable);
  $("#tower-auto .team-button").prop('disabled', inTeleop || !scoringAvailable);
  $(".counter button").prop('disabled', !scoringAvailable);
  $("#commit").prop('disabled', !commitAvailable);
  $("#fouls-button").prop('disabled', !scoringAvailable);
  $(".container").attr("data-in-teleop", inTeleop && scoringAvailable);
}

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  let realtimeScore = (alliance === "red") ? data.Red : data.Blue;
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const pos = i + 1;
    const level = score.TowerLevels[i];
    const isAuto = score.TowerIsAuto[i];

    let levelText = "None";
    if (level === 1) levelText = "Level 1";
    if (level === 2) levelText = "Level 2";
    if (level === 3) levelText = "Level 3";

    if (isAuto && level === 1) {
      $(`#tower-auto-${pos}-status`).text("Level 1");
      $(`#tower-auto-${pos}`).attr("data-selected", "true");
      $(`#tower-teleop-${pos}-status`).text("None");
      $(`#tower-teleop-${pos}`).attr("data-selected", "false");
    } else {
      $(`#tower-auto-${pos}-status`).text("None");
      $(`#tower-auto-${pos}`).attr("data-selected", "false");
      $(`#tower-teleop-${pos}-status`).text(levelText);
      $(`#tower-teleop-${pos}`).attr("data-selected", level > 0 ? "true" : "false");
    }
  }

  $(`#value-fuel-auto`).text(score.FuelAuto);
  $(`#value-fuel-transition`).text(score.FuelTransition);
  $(`#value-fuel-shift1`).text(score.FuelShift1);
  $(`#value-fuel-shift2`).text(score.FuelShift2);
  $(`#value-fuel-shift3`).text(score.FuelShift3);
  $(`#value-fuel-shift4`).text(score.FuelShift4);
  $(`#value-fuel-endgame`).text(score.FuelEndGame);
};

const cycleTowerAuto = function (index) {
  // Toggle between level 0 and level 1 (auto)
  let level = $(`#tower-auto-${index}-status`).text() === "Level 1" ? 0 : 1;
  websocket.send("tower", {
    TeamPosition: index,
    Level: level,
    IsAuto: level === 1
  });
}

const cycleTowerTeleop = function (index) {
  // Toggle between 0 -> 2 -> 3 -> 0
  let current = $(`#tower-teleop-${index}-status`).text();
  let level = 0;
  if (current === "None") level = 2;
  else if (current === "Level 2") level = 3;
  else level = 0;

  websocket.send("tower", {
    TeamPosition: index,
    Level: level,
    IsAuto: false
  });
}

const incrementCounter = function (id) {
  let shift = id.replace("fuel-", "");
  websocket.send("fuel", {Shift: shift, Adjustment: 1});
}

const decrementCounter = function (id) {
  let shift = id.replace("fuel-", "");
  websocket.send("fuel", {Shift: shift, Adjustment: -1});
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
  $(document).on("click", ".tower-auto-btn", function() {
    cycleTowerAuto(parseInt($(this).data("index")));
  });
  $(document).on("click", ".tower-teleop-btn", function() {
    cycleTowerTeleop(parseInt($(this).data("index")));
  });
  $(document).on("click", ".counter .plus", function() {
    incrementCounter($(this).closest(".counter").data("id"));
  });
  $(document).on("click", ".counter .minus", function() {
    decrementCounter($(this).closest(".counter").data("id"));
  });
  $(document).on("click", "#commit", function() {
    commitMatchScore();
  });
  $(document).on("click", "#fouls-button", function() {
    showFoulsDialog();
  });
  $(document).on("click", ".foul-button", function() {
    addFoul($(this).data("alliance"), $(this).data("type"));
  });

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function (event) { handleMatchLoad(event.data); },
    matchTime: function (event) { handleMatchTime(event.data); },
    realtimeScore: function (event) { handleRealtimeScore(event.data); },
    resetLocalState: function (event) { committed = false; updateUIMode(); },
  });
});

