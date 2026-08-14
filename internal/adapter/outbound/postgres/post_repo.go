package postgres

import (
	"context"
	"database/sql"
	"time"

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
	return r.attachEngagement(ctx, row.ID, row.PlayerID, row.PlayerName.String, teeTimeIDPtr(row.GroupID), row.Body, row.CreatedAt)
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
		detail, err := r.attachEngagement(ctx, row.ID, row.PlayerID, row.PlayerName.String, teeTimeIDPtr(row.GroupID), row.Body, row.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts[i] = *detail
	}
	return posts, nil
}

func (r *PostRepo) ListByGroupID(ctx context.Context, groupID int64, limit, offset int32) ([]entity.PostWithDetails, error) {
	rows, err := r.q.ListGroupPosts(ctx, sqlcgen.ListGroupPostsParams{
		GroupID: sql.NullInt64{Int64: groupID, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}
	posts := make([]entity.PostWithDetails, len(rows))
	for i, row := range rows {
		detail, err := r.attachEngagement(ctx, row.ID, row.PlayerID, row.PlayerName.String, teeTimeIDPtr(row.GroupID), row.Body, row.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts[i] = *detail
	}
	return posts, nil
}

func (r *PostRepo) Create(ctx context.Context, playerID int64, body string, groupID *int64) (int64, error) {
	row, err := r.q.CreatePost(ctx, sqlcgen.CreatePostParams{
		PlayerID: playerID,
		Body:     body,
		GroupID:  nullTeeTimeID(groupID),
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

func (r *PostRepo) DeleteByID(ctx context.Context, id int64) error {
	return r.q.DeletePostByID(ctx, id)
}

func (r *PostRepo) attachEngagement(ctx context.Context, id, playerID int64, playerName string, groupID *int64, body string, createdAt time.Time) (*entity.PostWithDetails, error) {
	reactions, err := r.q.ListReactionsByPostID(ctx, id)
	if err != nil {
		return nil, err
	}
	replies, err := r.q.ListRepliesByPostID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.PostWithDetails{
		ID:         id,
		PlayerID:   playerID,
		PlayerName: playerName,
		GroupID:    groupID,
		Body:       body,
		CreatedAt:  createdAt,
		Reactions:  mapReactions(reactions),
		Replies:    mapReplies(replies),
	}, nil
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
