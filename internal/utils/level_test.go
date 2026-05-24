package utils

import "testing"

func TestLevelForXP_Boundaries(t *testing.T) {
	cases := []struct {
		xp   int
		want int
	}{
		// Negative / zero — always level 0.
		{-1, 0},
		{0, 0},
		// Just below threshold for level 1 (100 XP needed).
		{99, 0},
		// Exactly at threshold — should be level 1.
		{100, 1},
		// Just below threshold for level 2 (400 XP needed).
		{399, 1},
		// Exactly at threshold — should be level 2.
		{400, 2},
		// Just below level 3 (900 XP).
		{899, 2},
		// Exactly at level 3.
		{900, 3},
		// Mid-range values.
		{1600, 4},
		{2500, 5},
	}

	for _, tc := range cases {
		got := LevelForXP(tc.xp)
		if got != tc.want {
			t.Errorf("LevelForXP(%d) = %d, want %d", tc.xp, got, tc.want)
		}
	}
}
