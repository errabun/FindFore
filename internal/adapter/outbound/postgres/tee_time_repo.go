package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type TeeTimeRepo struct {
	q *sqlcgen.Queries
}

func NewTeeTimeRepo(q *sqlcgen.Queries) *TeeTimeRepo {
	return &TeeTimeRepo{q: q}
}

func nullInt32Ptr(p *int32) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *p, Valid: true}
}

func int32Ptr(n sql.NullInt32) *int32 {
	if !n.Valid {
		return nil
	}
	v := n.Int32
	return &v
}

func mapTeeTime(
	id, courseID int64,
	startsAt time.Time,
	holes sql.NullString,
	status string,
	capacity, availableSlots, priceCents sql.NullInt32,
	currency sql.NullString,
) entity.TeeTime {
	return entity.TeeTime{
		ID:             id,
		CourseID:       courseID,
		StartsAt:       startsAt,
		Holes:          holes.String,
		Status:         status,
		Capacity:       int32Ptr(capacity),
		AvailableSlots: int32Ptr(availableSlots),
		PriceCents:     int32Ptr(priceCents),
		Currency:       currency.String,
	}
}

func (r *TeeTimeRepo) GetByID(ctx context.Context, id int64) (*entity.TeeTime, error) {
	row, err := r.q.GetTeeTimeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t := mapTeeTime(
		row.ID, row.CourseID, row.StartsAt, row.Holes, row.Status,
		row.Capacity, row.AvailableSlots, row.PriceCents, row.Currency,
	)
	return &t, nil
}

func (r *TeeTimeRepo) ListByCourseAndWindow(ctx context.Context, courseID int64, from, to time.Time) ([]entity.TeeTime, error) {
	rows, err := r.q.ListTeeTimesByCourseAndWindow(ctx, sqlcgen.ListTeeTimesByCourseAndWindowParams{
		CourseID:   courseID,
		StartsAt:   from,
		StartsAt_2: to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]entity.TeeTime, len(rows))
	for i, row := range rows {
		out[i] = mapTeeTime(
			row.ID, row.CourseID, row.StartsAt, row.Holes, row.Status,
			row.Capacity, row.AvailableSlots, row.PriceCents, row.Currency,
		)
	}
	return out, nil
}

func (r *TeeTimeRepo) GetByProviderExternalID(ctx context.Context, provider, externalID string) (*entity.TeeTime, error) {
	row, err := r.q.GetTeeTimeByProviderExternalID(ctx, sqlcgen.GetTeeTimeByProviderExternalIDParams{
		Provider:   provider,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}
	t := mapTeeTime(
		row.ID, row.CourseID, row.StartsAt, row.Holes, row.Status,
		row.Capacity, row.AvailableSlots, row.PriceCents, row.Currency,
	)
	return &t, nil
}

func (r *TeeTimeRepo) Create(ctx context.Context, t entity.TeeTime) (*entity.TeeTime, error) {
	status := t.Status
	if status == "" {
		status = entity.TeeTimeStatusAvailable
	}
	row, err := r.q.InsertTeeTime(ctx, sqlcgen.InsertTeeTimeParams{
		CourseID:       t.CourseID,
		StartsAt:       t.StartsAt,
		Holes:          sql.NullString{String: t.Holes, Valid: t.Holes != ""},
		Status:         status,
		Capacity:       nullInt32Ptr(t.Capacity),
		AvailableSlots: nullInt32Ptr(t.AvailableSlots),
		PriceCents:     nullInt32Ptr(t.PriceCents),
		Currency:       sql.NullString{String: t.Currency, Valid: t.Currency != ""},
	})
	if err != nil {
		return nil, err
	}
	out := mapTeeTime(
		row.ID, row.CourseID, row.StartsAt, row.Holes, row.Status,
		row.Capacity, row.AvailableSlots, row.PriceCents, row.Currency,
	)
	return &out, nil
}

func (r *TeeTimeRepo) UpdateCache(ctx context.Context, t entity.TeeTime) (*entity.TeeTime, error) {
	row, err := r.q.UpdateTeeTimeCache(ctx, sqlcgen.UpdateTeeTimeCacheParams{
		ID:             t.ID,
		Holes:          sql.NullString{String: t.Holes, Valid: t.Holes != ""},
		Status:         t.Status,
		Capacity:       nullInt32Ptr(t.Capacity),
		AvailableSlots: nullInt32Ptr(t.AvailableSlots),
		PriceCents:     nullInt32Ptr(t.PriceCents),
		Currency:       sql.NullString{String: t.Currency, Valid: t.Currency != ""},
		StartsAt:       t.StartsAt,
	})
	if err != nil {
		return nil, err
	}
	out := mapTeeTime(
		row.ID, row.CourseID, row.StartsAt, row.Holes, row.Status,
		row.Capacity, row.AvailableSlots, row.PriceCents, row.Currency,
	)
	return &out, nil
}

func (r *TeeTimeRepo) UpdateStatus(ctx context.Context, id int64, status string) (*entity.TeeTime, error) {
	row, err := r.q.UpdateTeeTimeStatus(ctx, sqlcgen.UpdateTeeTimeStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	out := mapTeeTime(
		row.ID, row.CourseID, row.StartsAt, row.Holes, row.Status,
		row.Capacity, row.AvailableSlots, row.PriceCents, row.Currency,
	)
	return &out, nil
}

func (r *TeeTimeRepo) GetProvider(ctx context.Context, provider, externalID string) (*entity.TeeTimeProvider, error) {
	row, err := r.q.GetTeeTimeProvider(ctx, sqlcgen.GetTeeTimeProviderParams{
		Provider:   provider,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}
	return &entity.TeeTimeProvider{
		ID:         row.ID,
		TeeTimeID:  row.TeeTimeID,
		Provider:   row.Provider,
		ExternalID: row.ExternalID,
	}, nil
}

func (r *TeeTimeRepo) LinkProvider(ctx context.Context, teeTimeID int64, provider, externalID string) error {
	existing, err := r.GetProvider(ctx, provider, externalID)
	if err == nil {
		if existing.TeeTimeID == teeTimeID {
			return nil
		}
		return entity.ErrProviderTeeTimeConflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = r.q.InsertTeeTimeProvider(ctx, sqlcgen.InsertTeeTimeProviderParams{
		TeeTimeID:  teeTimeID,
		Provider:   provider,
		ExternalID: externalID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			again, getErr := r.GetProvider(ctx, provider, externalID)
			if getErr != nil {
				return getErr
			}
			if again.TeeTimeID == teeTimeID {
				return nil
			}
			return entity.ErrProviderTeeTimeConflict
		}
		return err
	}
	return nil
}
