package entity

// CourseProvider tokens are app-owned lowercase strings (not PG ENUMs).
const (
	ProviderGolfCourseAPI = "golfcourseapi"
	ProviderLightspeed    = "lightspeed"
	ProviderForeUP        = "foreup"
	ProviderGolfNow       = "golfnow"
)

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

	// Provider / ExternalID are lookup metadata for find-or-create and search
	// results; they live in course_providers, not on the courses row.
	Provider   string
	ExternalID string
}
