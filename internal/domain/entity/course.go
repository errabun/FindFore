package entity

import "errors"

// CourseProvider tokens are app-owned lowercase strings (not PG ENUMs).
const (
	ProviderGolfCourseAPI = "golfcourseapi"
	ProviderLightspeed    = "lightspeed"
	ProviderForeUP        = "foreup"
	ProviderGolfNow       = "golfnow"
)

// Course is the canonical FindFore golf course. Provider external IDs live on CourseProvider.
type Course struct {
	ID        int64
	Name      string
	Street    string
	City      string
	State     string
	ZipCode   string
	Phone     string
	Cost      string
	Country   string
	Latitude  *float64
	Longitude *float64
	Timezone  string
}

// CourseProvider maps a vendor identity onto a canonical course.
type CourseProvider struct {
	ID         int64
	CourseID   int64
	Provider   string
	ExternalID string
}

// CourseSearchResult is a discovery hit: a course draft plus an optional provider link
// (not yet persisted until FindOrCreate).
type CourseSearchResult struct {
	Course     Course
	Provider   string
	ExternalID string
}

// ErrProviderCourseConflict is returned when (provider, external_id) is already
// linked to a different course.
var ErrProviderCourseConflict = errors.New("provider external id already linked to another course")
