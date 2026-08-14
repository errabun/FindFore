package feed

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

func (s *Service) Delete(ctx context.Context, postID, playerID int64) error {
	post, err := s.loadPost(ctx, postID)
	if err != nil {
		return err
	}
	if err := s.requireCanRead(ctx, post, playerID); err != nil {
		return err
	}
	if post.PlayerID != playerID && !s.canManageGroupPost(ctx, post, playerID) {
		return ErrPostNotFound
	}
	return s.posts.DeleteByID(ctx, postID)
}

func (s *Service) loadPost(ctx context.Context, postID int64) (*entity.PostWithDetails, error) {
	post, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return post, nil
}

func (s *Service) requireCanRead(ctx context.Context, post *entity.PostWithDetails, actorID int64) error {
	if post.GroupID == nil {
		return nil
	}
	return s.requireActiveMember(ctx, *post.GroupID, actorID)
}

func (s *Service) canManageGroupPost(ctx context.Context, post *entity.PostWithDetails, actorID int64) bool {
	if post.GroupID == nil || s.groups == nil {
		return false
	}
	m, err := s.groups.GetMembership(ctx, *post.GroupID, actorID)
	if err != nil {
		return false
	}
	return m.CanManage()
}

func (s *Service) requireActiveMember(ctx context.Context, groupID, actorID int64) error {
	if groupID <= 0 || actorID <= 0 {
		return ErrPostNotFound
	}
	if s.groups == nil {
		return ErrPostNotFound
	}
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPostNotFound
		}
		return err
	}
	m, err := s.groups.GetMembership(ctx, groupID, actorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPostNotFound
		}
		return err
	}
	if !m.IsActive() {
		return ErrPostNotFound
	}
	return nil
}
