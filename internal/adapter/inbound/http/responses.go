package httphandler

type CourseResponse struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Street    string   `json:"street"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	ZipCode   string   `json:"zip_code"`
	Phone     string   `json:"phone"`
	Cost      string   `json:"cost"`
	Country   string   `json:"country,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Timezone  string   `json:"timezone,omitempty"`
}

// CourseSearchResponse is a discovery hit. Provider identity is optional transport
// for find-or-create — not part of the canonical course representation.
type CourseSearchResponse struct {
	CourseResponse
	Provider   string `json:"provider,omitempty"`
	ExternalID string `json:"external_id,omitempty"`
}

type EventResponse struct {
	ID             int64   `json:"id"`
	CourseName     string  `json:"course_name"`
	Date           string  `json:"date"`
	TeeTime        string  `json:"tee_time"`
	OpenSpots      int32   `json:"open_spots"`
	NumberOfHoles  string  `json:"number_of_holes"`
	Private        bool    `json:"private"`
	HostName       string  `json:"host_name"`
	HostID         int32   `json:"host_id"`
	Accepted       []int64 `json:"accepted"`
	Declined       []int64 `json:"declined"`
	Pending        []int64 `json:"pending"`
	Closed         []int64 `json:"closed"`
	RemainingSpots int32   `json:"remaining_spots"`
	GroupID        *int64  `json:"group_id,omitempty"`
}

type PlayerResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Friends  []int64 `json:"friends"`
	Events   []int64 `json:"events"`
}

type LoginResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Friends  []int64 `json:"friends"`
	Events   []int64 `json:"events"`
	Token    string  `json:"token"`
}

type PlayerEventResponse struct {
	ID           int64  `json:"id"`
	PlayerID     int64  `json:"player_id"`
	EventID      int64  `json:"event_id"`
	InviteStatus string `json:"invite_status"`
}

type FriendshipResponse struct {
	ID          int64          `json:"id"`
	RequesterID int32          `json:"requester_id"`
	AddresseeID int32          `json:"addressee_id"`
	Status      string         `json:"status"`
	Requester   PlayerResponse `json:"requester"`
	Addressee   PlayerResponse `json:"addressee"`
}

type PostResponse struct {
	ID         int64              `json:"id"`
	PlayerID   int64              `json:"player_id"`
	PlayerName string             `json:"player_name"`
	Body       string             `json:"body"`
	CreatedAt  string             `json:"created_at"`
	Reactions  []ReactionResponse `json:"reactions"`
	Replies    []ReplyResponse    `json:"replies"`
}

type ReactionResponse struct {
	ID         int64  `json:"id"`
	PlayerID   int64  `json:"player_id"`
	PlayerName string `json:"player_name"`
	Emoji      string `json:"emoji"`
}

type ReplyResponse struct {
	ID         int64  `json:"id"`
	PlayerID   int64  `json:"player_id"`
	PlayerName string `json:"player_name"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
}
