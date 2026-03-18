package entity

import "time"

type Post struct {
	ID        int64
	PlayerID  int64
	Body      string
	CreatedAt time.Time
}

type PostWithDetails struct {
	ID         int64
	PlayerID   int64
	PlayerName string
	Body       string
	CreatedAt  time.Time
	Reactions  []Reaction
	Replies    []Reply
}

type Reaction struct {
	ID         int64
	PostID     int64
	PlayerID   int64
	PlayerName string
	Emoji      string
}

type Reply struct {
	ID         int64
	PostID     int64
	PlayerID   int64
	PlayerName string
	Body       string
	CreatedAt  time.Time
}
