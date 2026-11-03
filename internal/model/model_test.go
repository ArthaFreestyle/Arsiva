package model

import "testing"

func TestNewPageMetaData(t *testing.T) {
	cases := []struct {
		name               string
		page, size, total  int
		wantPage, wantSize int
		wantTotalPage      int
	}{
		{"exact multiple", 1, 10, 30, 1, 10, 3},
		{"partial last page", 2, 10, 31, 2, 10, 4},
		{"empty result", 1, 10, 0, 1, 10, 0},
		{"size zero falls back", 1, 0, 25, 1, DefaultPageSize, 3},
		{"size above max falls back", 1, MaxPageSize + 1, 25, 1, DefaultPageSize, 3},
		{"page below one clamps", 0, 5, 12, 1, 5, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewPageMetaData(tc.page, tc.size, tc.total)
			if got.Page != tc.wantPage || got.Size != tc.wantSize || got.TotalItem != tc.total || got.TotalPage != tc.wantTotalPage {
				t.Fatalf("got %+v, want page=%d size=%d total_item=%d total_page=%d",
					got, tc.wantPage, tc.wantSize, tc.total, tc.wantTotalPage)
			}
		})
	}
}
