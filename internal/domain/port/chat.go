package port

import "context"

// GroupChatSession is what the browser needs to connect to a group channel.
type GroupChatSession struct {
	APIKey      string
	Token       string
	ChannelType string
	ChannelID   string
	UserID      string
	UserName    string
}

// ChatProvider is the Stream (or future DIY) adapter at the edge.
type ChatProvider interface {
	APIKey() string
	IssueToken(ctx context.Context, playerID int64, name string) (string, error)
	EnsureGroupChannel(ctx context.Context, groupID int64, name string, members []GroupMember) error
}

// ChatService is the application use case: member-only group chat credentials.
type ChatService interface {
	GroupSession(ctx context.Context, actorID, groupID int64) (*GroupChatSession, error)
}
