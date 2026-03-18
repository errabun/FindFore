package postgres

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type ReactionRepo struct {
	q *sqlcgen.Queries
}

func NewReactionRepo(q *sqlcgen.Queries) *ReactionRepo {
	return &ReactionRepo{q: q}
}

func (r *ReactionRepo) ListByPostID(ctx context.Context, postID int64) ([]entity.Reaction, error) {
	rows, err := r.q.ListReactionsByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}
	return mapReactions(rows), nil
}

func (r *ReactionRepo) Find(ctx context.Context, postID, playerID int64, emoji string) (*entity.Reaction, error) {
	row, err := r.q.FindReaction(ctx, sqlcgen.FindReactionParams{
		PostID:   postID,
		PlayerID: playerID,
		Emoji:    emoji,
	})
	if err != nil {
		return nil, err
	}
	return &entity.Reaction{
		ID:       row.ID,
		PostID:   row.PostID,
		PlayerID: row.PlayerID,
		Emoji:    row.Emoji,
	}, nil
}

func (r *ReactionRepo) Create(ctx context.Context, postID, playerID int64, emoji string) (*entity.Reaction, error) {
	row, err := r.q.CreateReaction(ctx, sqlcgen.CreateReactionParams{
		PostID:   postID,
		PlayerID: playerID,
		Emoji:    emoji,
	})
	if err != nil {
		return nil, err
	}
	return &entity.Reaction{
		ID:       row.ID,
		PostID:   row.PostID,
		PlayerID: row.PlayerID,
		Emoji:    row.Emoji,
	}, nil
}

func (r *ReactionRepo) Delete(ctx context.Context, postID, playerID int64, emoji string) error {
	return r.q.DeleteReaction(ctx, sqlcgen.DeleteReactionParams{
		PostID:   postID,
		PlayerID: playerID,
		Emoji:    emoji,
	})
}
