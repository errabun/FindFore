package feed

import (
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var ErrPostNotFound = errors.New("post not found")

const (
	defaultListLimit = 20
	maxListLimit     = 50
)

// Service manages the community activity feed (posts, reactions, replies)
// and group-scoped posts. Community list is group_id IS NULL.
type Service struct {
	posts     port.PostRepository
	reactions port.ReactionRepository
	replies   port.ReplyRepository
	groups    port.GroupRepository
}

func NewService(posts port.PostRepository, reactions port.ReactionRepository, replies port.ReplyRepository, groups port.GroupRepository) *Service {
	return &Service{posts: posts, reactions: reactions, replies: replies, groups: groups}
}

func clampLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func clampOffset(offset int32) int32 {
	if offset < 0 {
		return 0
	}
	return offset
}
