// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package playoff

import (
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestDoubleEliminationInitial(t *testing.T) {
	finalMatchup, _, err := newDoubleEliminationBracket(8)
	assert.Nil(t, err)

	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)
	if assert.Equal(t, 19, len(matchSpecs)) {
		assertMatchSpecs(
			t,
			matchSpecs,
			[]expectedMatchSpec{
				{"Match 1", "M1", "Round 1 Upper", 1, "M1", true, false, "sf", 1, 1},
				{"Match 2", "M2", "Round 1 Upper", 2, "M2", true, false, "sf", 2, 1},
				{"Match 3", "M3", "Round 1 Upper", 3, "M3", true, false, "sf", 3, 1},
				{"Match 4", "M4", "Round 1 Upper", 4, "M4", true, false, "sf", 4, 1},
				{"Match 5", "M5", "Round 2 Lower", 5, "M5", true, false, "sf", 5, 1},
				{"Match 6", "M6", "Round 2 Lower", 6, "M6", true, false, "sf", 6, 1},
				{"Match 7", "M7", "Round 2 Upper", 7, "M7", true, false, "sf", 7, 1},
				{"Match 8", "M8", "Round 2 Upper", 8, "M8", true, false, "sf", 8, 1},
				{"Match 9", "M9", "Round 3 Lower", 9, "M9", true, false, "sf", 9, 1},
				{"Match 10", "M10", "Round 3 Lower", 10, "M10", true, false, "sf", 10, 1},
				{"Match 11", "M11", "Round 4 Upper", 11, "M11", true, false, "sf", 11, 1},
				{"Match 12", "M12", "Round 4 Lower", 12, "M12", true, false, "sf", 12, 1},
				{"Match 13", "M13", "Round 5 Lower", 13, "M13", true, false, "sf", 13, 1},
				{"Final 1", "F1", "", 14, "F", false, false, "f", 1, 1},
				{"Final 2", "F2", "", 15, "F", false, false, "f", 1, 2},
				{"Final 3", "F3", "", 16, "F", false, false, "f", 1, 3},
				{"Overtime 1", "O1", "", 17, "F", true, true, "f", 1, 4},
				{"Overtime 2", "O2", "", 18, "F", true, true, "f", 1, 5},
				{"Overtime 3", "O3", "", 19, "F", true, true, "f", 1, 6},
			},
		)
	}

	finalMatchup.update(map[int]playoffMatchResult{})
	assertMatchSpecAlliances(
		t,
		matchSpecs[0:4],
		[]expectedAlliances{
			{1, 8},
			{4, 5},
			{2, 7},
			{3, 6},
		},
	)
	for i := 4; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(
		t, matchGroups, "M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9", "M10", "M11", "M12", "M13", "F",
	)
}

func TestDoubleEliminationInitialWith6Alliances(t *testing.T) {
	finalMatchup, _, err := newDoubleEliminationBracket(6)
	assert.Nil(t, err)

	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)
	assert.Equal(t, 15, len(matchSpecs))

	// With 6 alliances, matches 1 and 3 are byes, so they are pruned.
	// But wait, M5 and M6 also become byes because one side is a selection source with ID > 6?
	// Let's check:
	// M1: 1 vs 8. W1=A1, L1=A8.
	// M3: 2 vs 7. W3=A2, L3=A7.
	// M7: W1 vs W2. W1 is pruned to A1. So M7 red source is A1.
	// M5: L1 vs L2. L1 is pruned to A8. So M5 red source is A8.
	// Since A8 > 6, M5 itself is pruned! W5=L2, L5=A8.
	// Similarly for M6: L3 is A7, so M6 is pruned. W6=L4, L6=A7.
	// Match 9: L7 vs W6. W6 is pruned to L4.
	// Match 10: L8 vs W5. W5 is pruned to L2.
	// Match groups collected should be: M2, M4, M7, M8, M9, M10, M11, M12, M13, F. (Missing: M1, M3, M5, M6)

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(
		t, matchGroups, "M2", "M4", "M7", "M8", "M9", "M10", "M11", "M12", "M13", "F",
	)

	// Check alliances for initial matches.
	// M2: 4 vs 5.
	// M4: 3 vs 6.
	// M7 red source is A1 (from pruned M1).
	// M7 blue source is W2 (from M2).
	// So initially M7 should be {1, 0}.
	assert.Equal(t, 10, len(matchGroups))
	finalMatchup.update(map[int]playoffMatchResult{})
	assert.Equal(t, 1, matchGroups["M7"].(*Matchup).RedAllianceId)
	assert.Equal(t, 0, matchGroups["M7"].(*Matchup).BlueAllianceId)
}

func TestDoubleEliminationInitialWith2Alliances(t *testing.T) {
	finalMatchup, _, err := newDoubleEliminationBracket(2)
	assert.Nil(t, err)

	matchSpecs, err := collectMatchSpecs(finalMatchup)
	assert.Nil(t, err)

	// With 2 alliances, only M11 and F remain. (M1-M10, M12, M13 are pruned)
	// Match groups collected: M11, F.
	// Total matches: 1 (M11) + 6 (F) = 7.
	assert.Equal(t, 7, len(matchSpecs))

	matchGroups, err := collectMatchGroups(finalMatchup)
	assert.Nil(t, err)
	assertMatchGroups(t, matchGroups, "M11", "F")

	finalMatchup.update(map[int]playoffMatchResult{})
	assert.Equal(t, 1, matchGroups["M11"].(*Matchup).RedAllianceId)
	assert.Equal(t, 2, matchGroups["M11"].(*Matchup).BlueAllianceId)
}

func TestDoubleEliminationErrors(t *testing.T) {
	_, _, err := newDoubleEliminationBracket(1)
	if assert.NotNil(t, err) {
		assert.Equal(t, "double-elimination bracket must have between 2 and 8 alliances", err.Error())
	}

	_, _, err = newDoubleEliminationBracket(9)
	if assert.NotNil(t, err) {
		assert.Equal(t, "double-elimination bracket must have between 2 and 8 alliances", err.Error())
	}
}

func TestDoubleEliminationProgression(t *testing.T) {
	playoffTournament, err := NewPlayoffTournament(model.DoubleEliminationPlayoff, 8)
	assert.Nil(t, err)
	finalMatchup := playoffTournament.FinalMatchup()
	matchSpecs := playoffTournament.matchSpecs
	matchGroups := playoffTournament.MatchGroups()
	playoffMatchResults := map[int]playoffMatchResult{}

	assertMatchupOutcome(t, matchGroups["M1"], "", "")

	playoffMatchResults[1] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[4:7], []expectedAlliances{{8, 0}, {0, 0}, {1, 0}})
	for i := 7; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}
	assertMatchupOutcome(
		t, matchGroups["M1"], "Advances to Match 7 &ndash; Round 2 Upper", "Advances to Match 5 &ndash; Round 2 Lower",
	)

	// Reverse a previous outcome.
	playoffMatchResults[1] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[4:7], []expectedAlliances{{1, 0}, {0, 0}, {8, 0}})
	for i := 7; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}
	assertMatchupOutcome(
		t, matchGroups["M1"], "Advances to Match 5 &ndash; Round 2 Lower", "Advances to Match 7 &ndash; Round 2 Upper",
	)

	playoffMatchResults[2] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[4:7], []expectedAlliances{{1, 5}, {0, 0}, {8, 4}})
	for i := 7; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}
	assertMatchupOutcome(
		t, matchGroups["M2"], "Advances to Match 7 &ndash; Round 2 Upper", "Advances to Match 5 &ndash; Round 2 Lower",
	)

	playoffMatchResults[3] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[5:8], []expectedAlliances{{2, 0}, {8, 4}, {7, 0}})
	for i := 8; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}
	assertMatchupOutcome(
		t, matchGroups["M3"], "Advances to Match 6 &ndash; Round 2 Lower", "Advances to Match 8 &ndash; Round 2 Upper",
	)

	playoffMatchResults[4] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[5:8], []expectedAlliances{{2, 6}, {8, 4}, {7, 3}})
	for i := 8; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	playoffMatchResults[5] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[8:10], []expectedAlliances{{0, 0}, {0, 5}})
	for i := 10; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}
	assertMatchupOutcome(t, matchGroups["M5"], "Eliminated", "Advances to Match 10 &ndash; Round 3 Lower")

	playoffMatchResults[6] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[8:10], []expectedAlliances{{0, 2}, {0, 5}})
	for i := 10; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	// Score a perfect tie; no alliance should advance until the match is replayed.
	playoffMatchResults[7] = playoffMatchResult{game.TieMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[8:10], []expectedAlliances{{0, 2}, {0, 5}})
	for i := 10; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	playoffMatchResults[7] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[8:11], []expectedAlliances{{8, 2}, {0, 5}, {4, 0}})
	for i := 11; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	playoffMatchResults[8] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[8:11], []expectedAlliances{{8, 2}, {7, 5}, {4, 3}})
	for i := 11; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	// Score two matches at the same time.
	playoffMatchResults[9] = playoffMatchResult{game.RedWonMatch}
	playoffMatchResults[10] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[11:12], []expectedAlliances{{7, 8}})
	for i := 12; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{0, 0}})
	}

	playoffMatchResults[11] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[12:13], []expectedAlliances{{3, 0}})
	finalMatchup.update(playoffMatchResults)
	for i := 13; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{4, 0}})
	}
	assertMatchupOutcome(
		t, matchGroups["M11"], "Advances to Final 1", "Advances to Match 13 &ndash; Round 5 Lower",
	)

	playoffMatchResults[12] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[12:13], []expectedAlliances{{3, 7}})
	for i := 13; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{4, 0}})
	}

	playoffMatchResults[13] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	for i := 13; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{4, 3}})
	}
	assertMatchupOutcome(t, matchGroups["M13"], "Advances to Final 1", "Eliminated")

	// Unscore the previous match.
	delete(playoffMatchResults, 13)
	finalMatchup.update(playoffMatchResults)
	assertMatchSpecAlliances(t, matchSpecs[12:13], []expectedAlliances{{3, 7}})
	for i := 13; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{4, 0}})
	}
	assertMatchupOutcome(t, matchGroups["M13"], "", "")

	playoffMatchResults[13] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	for i := 13; i < 19; i++ {
		assertMatchSpecAlliances(t, matchSpecs[i:i+1], []expectedAlliances{{4, 7}})
	}
	assertMatchupOutcome(t, matchGroups["M13"], "Eliminated", "Advances to Final 1")

	playoffMatchResults[14] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[15] = playoffMatchResult{game.RedWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[16] = playoffMatchResult{game.TieMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[17] = playoffMatchResult{game.TieMatch}
	finalMatchup.update(playoffMatchResults)
	assert.False(t, finalMatchup.IsComplete())
	assert.Equal(t, 0, finalMatchup.WinningAllianceId())
	assert.Equal(t, 0, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "", "")

	playoffMatchResults[18] = playoffMatchResult{game.BlueWonMatch}
	finalMatchup.update(playoffMatchResults)
	assert.True(t, finalMatchup.IsComplete())
	assert.Equal(t, 7, finalMatchup.WinningAllianceId())
	assert.Equal(t, 4, finalMatchup.LosingAllianceId())
	assertMatchupOutcome(t, matchGroups["F"], "Tournament Finalist", "Tournament Winner")
}
