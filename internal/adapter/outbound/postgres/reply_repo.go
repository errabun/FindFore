package postgres

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type ReplyRepo struct {
	q *sqlcgen.Queries
}

func NewReplyRepo(q *sqlcgen.Queries) *ReplyRepo {
	return &ReplyRepo{q: q}
}

func (r *ReplyRepo) ListByPostID(ctx context.Context, postID int64) ([]entity.Reply, error) {
	rows, err := r.q.ListRepliesByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}
	return mapReplies(rows), nil
}

func (r *ReplyRepo) Create(ctx context.Context, postID, playerID int64, body string) (*entity.Reply, error) {
	row, err := r.q.CreateReply(ctx, sqlcgen.CreateReplyParams{
		PostID:   postID,
		PlayerID: playerID,
		Body:     body,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the full row with player name via JOIN
	full, err := r.q.GetReplyByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return &entity.Reply{
		ID:         full.ID,
		PostID:     full.PostID,
		PlayerID:   full.PlayerID,
		PlayerName: full.PlayerName.String,
		Body:       full.Body,
		CreatedAt:  full.CreatedAt,
	}, nil
}

func (r *ReplyRepo) Delete(ctx context.Context, id, playerID int64) error {
	return r.q.DeleteReply(ctx, sqlcgen.DeleteReplyParams{
		ID:       id,
		PlayerID: playerID,
	})
}
