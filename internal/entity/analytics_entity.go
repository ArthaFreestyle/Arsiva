package entity

import "time"

// UserRoleCount holds active-user counts per role.
type UserRoleCount struct {
	SuperAdmin int
	Guru       int
	Member     int
}

// ContentCount holds total/published counts for every content table.
// Draft counts are derived (total - published) by the usecase.
type ContentCount struct {
	KuisTotal        int
	KuisPublished    int
	CeritaTotal      int
	CeritaPublished  int
	PuzzleTotal      int
	PuzzlePublished  int
	ArtikelTotal     int
	ArtikelPublished int
}

// ContentTypeCount holds a count per content type.
type ContentTypeCount struct {
	Kuis    int
	Cerita  int
	Puzzle  int
	Artikel int
}

// UserGrowthBucket is the number of users created inside one time bucket.
// Bucket is the bucket start as Jakarta wall-clock time; only its date part
// is meaningful.
type UserGrowthBucket struct {
	Bucket time.Time
	UserRoleCount
}

// ContentGrowthBucket is the number of content rows created inside one time bucket.
type ContentGrowthBucket struct {
	Bucket time.Time
	ContentTypeCount
}
