// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific audience sound timings.

package game

type MatchSound struct {
	Name          string
	FileExtension string
	MatchTimeSec  float64
}

// List of sounds and how many seconds into the match they are played. A negative time indicates that the sound can only
// be triggered explicitly.
var MatchSounds []*MatchSound

func UpdateMatchSounds() {
	MatchSounds = []*MatchSound{
		{
			"start",
			"wav",
			float64(MatchTiming.WarmupDurationSec),
		},
		{
			"end",
			"wav",
			GetDurationToAutoEnd().Seconds(),
		},
		{
			"resume",
			"wav",
			GetDurationToTeleopStart().Seconds(),
		},
		{
			"warning_sonar",
			"wav",
			GetDurationToTeleopEnd().Seconds() - float64(MatchTiming.WarningRemainingDurationSec),
		},
		{
			"end",
			"wav",
			GetDurationToTeleopEnd().Seconds(),
		},
		{
			"abort",
			"wav",
			-1,
		},
		{
			"match_result",
			"wav",
			-1,
		},
		{
			"pick_clock",
			"wav",
			-1,
		},
		{
			"pick_clock_expired",
			"wav",
			-1,
		},
	}
}
