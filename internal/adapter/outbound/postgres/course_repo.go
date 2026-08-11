package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type CourseRepo struct {
	q *sqlcgen.Queries
}

func NewCourseRepo(q *sqlcgen.Queries) *CourseRepo {
	return &CourseRepo{q: q}
}

func nullFloat(p *float64) sql.NullFloat64 {
	if p == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *p, Valid: true}
}

func floatPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

func mapCourseRow(
	id int64,
	name, street, city, state, zipCode, phone, cost, country, timezone sql.NullString,
	lat, lng sql.NullFloat64,
) entity.Course {
	return entity.Course{
		ID:        id,
		Name:      name.String,
		Street:    street.String,
		City:      city.String,
		State:     state.String,
		ZipCode:   zipCode.String,
		Phone:     phone.String,
		Cost:      cost.String,
		Country:   country.String,
		Latitude:  floatPtr(lat),
		Longitude: floatPtr(lng),
		Timezone:  timezone.String,
	}
}

func (r *CourseRepo) List(ctx context.Context) ([]entity.Course, error) {
	rows, err := r.q.ListCourses(ctx)
	if err != nil {
		return nil, err
	}
	courses := make([]entity.Course, len(rows))
	for i, row := range rows {
		courses[i] = mapCourseRow(
			row.ID, row.Name, row.Street, row.City, row.State, row.ZipCode, row.Phone, row.Cost,
			row.Country, row.Timezone, row.Latitude, row.Longitude,
		)
	}
	return courses, nil
}

func (r *CourseRepo) GetByID(ctx context.Context, id int64) (*entity.Course, error) {
	row, err := r.q.GetCourseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c := mapCourseRow(
		row.ID, row.Name, row.Street, row.City, row.State, row.ZipCode, row.Phone, row.Cost,
		row.Country, row.Timezone, row.Latitude, row.Longitude,
	)
	return &c, nil
}

func (r *CourseRepo) GetByNameAndCity(ctx context.Context, name, city string) (*entity.Course, error) {
	row, err := r.q.GetCourseByNameAndCity(ctx, sqlcgen.GetCourseByNameAndCityParams{
		Name: sql.NullString{String: name, Valid: true},
		City: sql.NullString{String: city, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	c := mapCourseRow(
		row.ID, row.Name, row.Street, row.City, row.State, row.ZipCode, row.Phone, row.Cost,
		row.Country, row.Timezone, row.Latitude, row.Longitude,
	)
	return &c, nil
}

func (r *CourseRepo) GetByProviderExternalID(ctx context.Context, provider, externalID string) (*entity.Course, error) {
	row, err := r.q.GetCourseByProviderExternalID(ctx, sqlcgen.GetCourseByProviderExternalIDParams{
		Provider:   provider,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}
	c := mapCourseRow(
		row.ID, row.Name, row.Street, row.City, row.State, row.ZipCode, row.Phone, row.Cost,
		row.Country, row.Timezone, row.Latitude, row.Longitude,
	)
	return &c, nil
}

func (r *CourseRepo) Create(ctx context.Context, c entity.Course) (*entity.Course, error) {
	country := c.Country
	if country == "" {
		country = "US"
	}
	row, err := r.q.CreateCourse(ctx, sqlcgen.CreateCourseParams{
		Name:      sql.NullString{String: c.Name, Valid: true},
		Street:    sql.NullString{String: c.Street, Valid: true},
		City:      sql.NullString{String: c.City, Valid: true},
		State:     sql.NullString{String: c.State, Valid: true},
		ZipCode:   sql.NullString{String: c.ZipCode, Valid: true},
		Phone:     sql.NullString{String: c.Phone, Valid: true},
		Cost:      sql.NullString{String: c.Cost, Valid: true},
		Country:   sql.NullString{String: country, Valid: true},
		Latitude:  nullFloat(c.Latitude),
		Longitude: nullFloat(c.Longitude),
		Timezone:  sql.NullString{String: c.Timezone, Valid: c.Timezone != ""},
	})
	if err != nil {
		return nil, err
	}
	out := mapCourseRow(
		row.ID, row.Name, row.Street, row.City, row.State, row.ZipCode, row.Phone, row.Cost,
		row.Country, row.Timezone, row.Latitude, row.Longitude,
	)
	return &out, nil
}

func (r *CourseRepo) GetProvider(ctx context.Context, provider, externalID string) (*entity.CourseProvider, error) {
	row, err := r.q.GetCourseProvider(ctx, sqlcgen.GetCourseProviderParams{
		Provider:   provider,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}
	return &entity.CourseProvider{
		ID:         row.ID,
		CourseID:   row.CourseID,
		Provider:   row.Provider,
		ExternalID: row.ExternalID,
	}, nil
}

func (r *CourseRepo) LinkProvider(ctx context.Context, courseID int64, provider, externalID string) error {
	existing, err := r.GetProvider(ctx, provider, externalID)
	if err == nil {
		if existing.CourseID == courseID {
			return nil
		}
		return entity.ErrProviderCourseConflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = r.q.InsertCourseProvider(ctx, sqlcgen.InsertCourseProviderParams{
		CourseID:   courseID,
		Provider:   provider,
		ExternalID: externalID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			// Concurrent insert — re-check ownership.
			again, getErr := r.GetProvider(ctx, provider, externalID)
			if getErr != nil {
				return getErr
			}
			if again.CourseID == courseID {
				return nil
			}
			return entity.ErrProviderCourseConflict
		}
		return err
	}
	return nil
}

