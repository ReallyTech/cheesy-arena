// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RobotsBypassed [3]bool
	FuelAuto       int
	FuelTransition int
	FuelShift1     int
	FuelShift2     int
	FuelShift3     int
	FuelShift4     int
	FuelEndGame    int
	TowerLevels    [3]int
	TowerIsAuto    [3]bool
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

	// HUB Activation Logic
	// Always Active: Auto, Transition Shift, End Game
	// Shifts 1-4: Alternating based on Auto Winner

	selfAutoFuel := score.FuelAuto
	oppAutoFuel := opponentScore.FuelAuto

	selfActiveShift1 := true
	if selfAutoFuel > oppAutoFuel {
		selfActiveShift1 = false // Auto winner is inactive in Shift 1
	} else if oppAutoFuel > selfAutoFuel {
		selfActiveShift1 = true // Auto loser is active in Shift 1
	} else {
		// Tied: Randomly select one. We use MatchId and IsRed for a deterministic "random" choice.
		if (matchId%2 == 0) == isRed {
			selfActiveShift1 = false
		} else {
			selfActiveShift1 = true
		}
	}

	selfActiveShift2 := !selfActiveShift1
	selfActiveShift3 := selfActiveShift1
	selfActiveShift4 := !selfActiveShift1

	// Calculate Fuel points
	summary.FuelPoints = score.FuelAuto + score.FuelTransition + score.FuelEndGame
	summary.TotalFuel = score.FuelAuto + score.FuelTransition + score.FuelShift1 + score.FuelShift2 + score.FuelShift3 + score.FuelShift4 + score.FuelEndGame
	if selfActiveShift1 {
		summary.FuelPoints += score.FuelShift1
	}
	if selfActiveShift2 {
		summary.FuelPoints += score.FuelShift2
	}
	if selfActiveShift3 {
		summary.FuelPoints += score.FuelShift3
	}
	if selfActiveShift4 {
		summary.FuelPoints += score.FuelShift4
	}

	// Calculate Tower points
	autoLevel1Count := 0
	for i, level := range score.TowerLevels {
		if level == 1 && score.TowerIsAuto[i] {
			autoLevel1Count++
		}
	}
	if autoLevel1Count > 2 {
		autoLevel1Count = 2
	}
	summary.AutoFuelPoints = score.FuelAuto
	summary.AutoTowerPoints = autoLevel1Count * 15
	summary.AutoPoints = summary.AutoFuelPoints + summary.AutoTowerPoints
	summary.TowerPoints = summary.AutoTowerPoints
	summary.TotalTowers = 0

	for i, level := range score.TowerLevels {
		if level > 0 {
			summary.TotalTowers++
		}
		if score.TowerIsAuto[i] {
			continue // Already counted in AutoTowerPoints if L1
		}
		switch level {
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
	return score.FuelAuto + score.FuelTransition + score.FuelShift1 + score.FuelShift2 + score.FuelShift3 + score.FuelShift4 + score.FuelEndGame
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
		score.FuelTransition != other.FuelTransition ||
		score.FuelShift1 != other.FuelShift1 ||
		score.FuelShift2 != other.FuelShift2 ||
		score.FuelShift3 != other.FuelShift3 ||
		score.FuelShift4 != other.FuelShift4 ||
		score.FuelEndGame != other.FuelEndGame ||
		score.TowerLevels != other.TowerLevels ||
		score.TowerIsAuto != other.TowerIsAuto ||
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
