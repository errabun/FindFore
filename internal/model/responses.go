package model

type CourseResponse struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Phone   string `json:"phone"`
	Cost    string `json:"cost"`
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
	ID         int64          `json:"id"`
	FollowerID int32          `json:"follower_id"`
	FolloweeID int32          `json:"followee_id"`
	Follower   PlayerResponse `json:"follower"`
	Followee   PlayerResponse `json:"followee"`
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

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors []ErrorDetail `json:"errors"`
}
