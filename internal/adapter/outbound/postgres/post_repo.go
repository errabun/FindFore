package postgres

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PostRepo struct {
	q *sqlcgen.Queries
}

func NewPostRepo(q *sqlcgen.Queries) *PostRepo {
	return &PostRepo{q: q}
}

func (r *PostRepo) GetByID(ctx context.Context, id int64) (*entity.PostWithDetails, error) {
	row, err := r.q.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reactions, err := r.q.ListReactionsByPostID(ctx, id)
	if err != nil {
		return nil, err
	}

	replies, err := r.q.ListRepliesByPostID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &entity.PostWithDetails{
		ID:         row.ID,
		PlayerID:   row.PlayerID,
		PlayerName: row.PlayerName.String,
		Body:       row.Body,
		CreatedAt:  row.CreatedAt,
		Reactions:  mapReactions(reactions),
		Replies:    mapReplies(replies),
	}, nil
}

func (r *PostRepo) List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error) {
	rows, err := r.q.ListPosts(ctx, sqlcgen.ListPostsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	posts := make([]entity.PostWithDetails, len(rows))
	for i, row := range rows {
		reactions, err := r.q.ListReactionsByPostID(ctx, row.ID)
		if err != nil {
			return nil, err
		}

		replies, err := r.q.ListRepliesByPostID(ctx, row.ID)
		if err != nil {
			return nil, err
		}

		posts[i] = entity.PostWithDetails{
			ID:         row.ID,
			PlayerID:   row.PlayerID,
			PlayerName: row.PlayerName.String,
			Body:       row.Body,
			CreatedAt:  row.CreatedAt,
			Reactions:  mapReactions(reactions),
			Replies:    mapReplies(replies),
		}
	}
	return posts, nil
}

func (r *PostRepo) Create(ctx context.Context, playerID int64, body string) (int64, error) {
	row, err := r.q.CreatePost(ctx, sqlcgen.CreatePostParams{
		PlayerID: playerID,
		Body:     body,
	})
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *PostRepo) Delete(ctx context.Context, id, playerID int64) error {
	return r.q.DeletePost(ctx, sqlcgen.DeletePostParams{
		ID:       id,
		PlayerID: playerID,
	})
}

func mapReactions(rows []sqlcgen.ListReactionsByPostIDRow) []entity.Reaction {
	reactions := make([]entity.Reaction, len(rows))
	for i, row := range rows {
		reactions[i] = entity.Reaction{
			ID:         row.ID,
			PostID:     row.PostID,
			PlayerID:   row.PlayerID,
			PlayerName: row.PlayerName.String,
			Emoji:      row.Emoji,
		}
	}
	return reactions
}

func mapReplies(rows []sqlcgen.ListRepliesByPostIDRow) []entity.Reply {
	replies := make([]entity.Reply, len(rows))
	for i, row := range rows {
		replies[i] = entity.Reply{
			ID:         row.ID,
			PostID:     row.PostID,
			PlayerID:   row.PlayerID,
			PlayerName: row.PlayerName.String,
			Body:       row.Body,
			CreatedAt:  row.CreatedAt,
		}
	}
	return replies
}
