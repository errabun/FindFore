package postgres

import (
	"context"
	"database/sql"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PlayerRepo struct {
	q *sqlcgen.Queries
}

func NewPlayerRepo(q *sqlcgen.Queries) *PlayerRepo {
	return &PlayerRepo{q: q}
}

func (r *PlayerRepo) List(ctx context.Context) ([]entity.Player, error) {
	rows, err := r.q.ListPlayers(ctx)
	if err != nil {
		return nil, err
	}
	players := make([]entity.Player, len(rows))
	for i, row := range rows {
		players[i] = entity.Player{
			ID:       row.ID,
			Name:     row.Name.String,
			Phone:    row.Phone.String,
			Email:    row.Email.String,
			Username: row.Username.String,
		}
	}
	return players, nil
}

func (r *PlayerRepo) GetByID(ctx context.Context, id int64) (*entity.Player, error) {
	row, err := r.q.GetPlayerByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.Player{
		ID:       row.ID,
		Name:     row.Name.String,
		Phone:    row.Phone.String,
		Email:    row.Email.String,
		Username: row.Username.String,
	}, nil
}

func (r *PlayerRepo) GetByEmail(ctx context.Context, email string) (*entity.Player, error) {
	row, err := r.q.GetPlayerByEmail(ctx, sql.NullString{String: email, Valid: true})
	if err != nil {
		return nil, err
	}
	return &entity.Player{
		ID:             row.ID,
		Name:           row.Name.String,
		Phone:          row.Phone.String,
		Email:          row.Email.String,
		Username:       row.Username.String,
		PasswordDigest: row.PasswordDigest.String,
	}, nil
}

func (r *PlayerRepo) GetByUsername(ctx context.Context, username string) (*entity.Player, error) {
	row, err := r.q.GetPlayerByUsername(ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		return nil, err
	}
	return &entity.Player{
		ID:             row.ID,
		Name:           row.Name.String,
		Phone:          row.Phone.String,
		Email:          row.Email.String,
		Username:       row.Username.String,
		PasswordDigest: row.PasswordDigest.String,
	}, nil
}

func (r *PlayerRepo) Create(ctx context.Context, p entity.Player) (*entity.Player, error) {
	row, err := r.q.CreatePlayer(ctx, sqlcgen.CreatePlayerParams{
		Name:           sql.NullString{String: p.Name, Valid: true},
		Phone:          sql.NullString{String: p.Phone, Valid: true},
		Email:          sql.NullString{String: p.Email, Valid: true},
		Username:       sql.NullString{String: p.Username, Valid: true},
		PasswordDigest: sql.NullString{String: p.PasswordDigest, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.Player{
		ID:       row.ID,
		Name:     row.Name.String,
		Phone:    row.Phone.String,
		Email:    row.Email.String,
		Username: row.Username.String,
	}, nil
}

func (r *PlayerRepo) ListIDsExcept(ctx context.Context, excludeID int64) ([]int64, error) {
	return r.q.ListPlayersExceptHost(ctx, excludeID)
}
