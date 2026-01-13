// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummaryDetermineMatchStatus(t *testing.T) {
	redScoreSummary := &ScoreSummary{Score: 10}
	blueScoreSummary := &ScoreSummary{Score: 10}
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.Score = 11
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.Score = 12
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.Score = 12
	redScoreSummary.FuelPoints = 11
	redScoreSummary.TowerPoints = 11
	redScoreSummary.AutoFuelPoints = 11
	redScoreSummary.AutoTowerPoints = 11
	blueScoreSummary.FuelPoints = 10
	blueScoreSummary.TowerPoints = 10
	blueScoreSummary.AutoFuelPoints = 10
	blueScoreSummary.AutoTowerPoints = 10
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	// Fuel tiebreaker
	blueScoreSummary.FuelPoints = 12
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.FuelPoints = 12
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true)) // Tower tiebreaker: 11 vs 10

	// Tower tiebreaker
	blueScoreSummary.TowerPoints = 12
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.TowerPoints = 12
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true)) // Auto Fuel tiebreaker: 11 vs 10

	// Auto Fuel tiebreaker
	blueScoreSummary.AutoFuelPoints = 12
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.AutoFuelPoints = 12
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true)) // Auto Tower tiebreaker: 11 vs 10

	// Auto Tower tiebreaker
	blueScoreSummary.AutoTowerPoints = 12
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
	redScoreSummary.AutoTowerPoints = 12
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
}
