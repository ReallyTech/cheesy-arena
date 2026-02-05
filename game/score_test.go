// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore, true, 1)
	assert.Equal(t, 170, redSummary.FuelPoints)
	assert.Equal(t, 45, redSummary.TowerPoints)
	assert.Equal(t, 35, redSummary.AutoPoints) // FuelAuto (20) + TowerAuto (15)
	assert.Equal(t, 215, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 215, redSummary.Score)
	assert.Equal(t, true, redSummary.FuelEnergizedRankingPoint)
	assert.Equal(t, false, redSummary.FuelSuperchargedRankingPoint)
	assert.Equal(t, true, redSummary.TowerTraversalRankingPoint)
	assert.Equal(t, 2, redSummary.BonusRankingPoints)

	blueSummary := blueScore.Summarize(redScore, false, 1)
	assert.Equal(t, 105, blueSummary.FuelPoints)
	assert.Equal(t, 80, blueSummary.TowerPoints)
	assert.Equal(t, 40, blueSummary.AutoPoints) // FuelAuto (10) + TowerAuto (30)
	assert.Equal(t, 185, blueSummary.MatchPoints)
	assert.Equal(t, 20, blueSummary.FoulPoints) // 5 (Minor) + 15 (Major)
	assert.Equal(t, 205, blueSummary.Score)
	assert.Equal(t, true, blueSummary.FuelEnergizedRankingPoint)
	assert.Equal(t, false, blueSummary.FuelSuperchargedRankingPoint)
	assert.Equal(t, true, blueSummary.TowerTraversalRankingPoint)
	assert.Equal(t, 2, blueSummary.BonusRankingPoints)
}

func TestScorePlayoffDq(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redScore.PlayoffDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore, true, 1).Score)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	score2 = TestScore1()
	score2.RobotsBypassed[0] = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.FuelAuto += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.TowerLevels[1] = 1
	score2.TowerAuto[1] = true
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
