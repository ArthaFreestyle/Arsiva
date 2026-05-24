package utils

import "math"

// levelXpFactor is K in the formula level = floor(sqrt(total_xp / K)).
// XP required to reach level n is K * n².
const levelXpFactor = 100

// LevelForXP returns the level that corresponds to totalXp accumulated XP.
// Formula: floor(sqrt(total_xp / 100)). Negative input is clamped to level 0.
func LevelForXP(totalXp int) int {
	if totalXp <= 0 {
		return 0
	}
	return int(math.Sqrt(float64(totalXp) / float64(levelXpFactor)))
}
