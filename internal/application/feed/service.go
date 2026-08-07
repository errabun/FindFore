package feed

import (
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// Service manages the community activity feed (posts, reactions, replies).
type Service struct {
	posts     port.PostRepository
	reactions port.ReactionRepository
	replies   port.ReplyRepository
}

func NewService(posts port.PostRepository, reactions port.ReactionRepository, replies port.ReplyRepository) *Service {
	return &Service{posts: posts, reactions: reactions, replies: replies}
}
