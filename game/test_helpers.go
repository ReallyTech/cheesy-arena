// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{1, true, 254, 1},
		{2, false, 254, 2},
	}
	return &Score{
		RobotsBypassed: [3]bool{false, false, true},
		FuelAuto:       20,
		FuelTeleop:     150,
		TowerLevels:    [3]int{1, 2, 0},
		TowerAuto:      [3]bool{true, false, false},
		Fouls:          fouls,
		PlayoffDq:      false,
	}
}

func TestScore2() *Score {
	return &Score{
		RobotsBypassed: [3]bool{false, false, false},
		FuelAuto:       10,
		FuelTeleop:     95,
		TowerLevels:    [3]int{1, 1, 3},
		TowerAuto:      [3]bool{true, true, false},
		Fouls:          []Foul{},
		PlayoffDq:      false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 12, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 23, 0.1114, 1, 3, 2, 0, 10}}
}
