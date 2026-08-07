package feed

import "context"

func (s *Service) Delete(ctx context.Context, postID, playerID int64) error {
	return s.posts.Delete(ctx, postID, playerID)
}
