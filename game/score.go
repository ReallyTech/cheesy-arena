// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RobotsBypassed [3]bool
	FuelAuto       int
	FuelTeleop     int
	TowerLevels    [3]int
	TowerAuto      [3]bool
	Fouls          []Foul
	PlayoffDq      bool
}

// Game-specific settings that can be changed via the settings.
var FuelEnergizedThreshold = 100
var FuelSuperchargedThreshold = 360
var TowerTraversalThreshold = 2

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score, isRed bool, matchId int) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate Fuel points
	summary.FuelPoints = score.FuelAuto + score.FuelTeleop
	summary.TotalFuel = score.FuelAuto + score.FuelTeleop

	// Calculate Tower points
	autoClimbCount := 0
	for _, auto := range score.TowerAuto {
		if auto {
			autoClimbCount++
		}
	}
	if autoClimbCount > 2 {
		autoClimbCount = 2
	}
	summary.AutoFuelPoints = score.FuelAuto
	summary.AutoTowerPoints = autoClimbCount * 15
	summary.AutoPoints = summary.AutoFuelPoints + summary.AutoTowerPoints
	summary.TowerPoints = summary.AutoTowerPoints
	summary.TotalTowers = 0

	for _, level := range score.TowerLevels {
		if level > 0 {
			summary.TotalTowers++
		}
		switch level {
		case 1:
			summary.TowerPoints += 10
		case 2:
			summary.TowerPoints += 20
		case 3:
			summary.TowerPoints += 30
		}
	}

	summary.MatchPoints = summary.FuelPoints + summary.TowerPoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Ranking Points
	if summary.TotalFuel >= FuelEnergizedThreshold {
		summary.FuelEnergizedRankingPoint = true
		summary.BonusRankingPoints++
	}
	if summary.TotalFuel >= FuelSuperchargedThreshold {
		summary.FuelSuperchargedRankingPoint = true
		summary.BonusRankingPoints++
	}
	if summary.TotalTowers >= TowerTraversalThreshold {
		summary.TowerTraversalRankingPoint = true
		summary.BonusRankingPoints++
	}

	return summary
}

func (score *Score) FuelTotal() int {
	return score.FuelAuto + score.FuelTeleop
}

func (score *Score) TowerLevelCount(level int) int {
	count := 0
	for _, l := range score.TowerLevels {
		if l == level {
			count++
		}
	}
	return count
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.RobotsBypassed != other.RobotsBypassed ||
		score.FuelAuto != other.FuelAuto ||
		score.FuelTeleop != other.FuelTeleop ||
		score.TowerLevels != other.TowerLevels ||
		score.TowerAuto != other.TowerAuto ||
		score.PlayoffDq != other.PlayoffDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
