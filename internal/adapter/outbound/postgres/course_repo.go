package postgres

import (
	"context"
	"database/sql"

	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type CourseRepo struct {
	q *sqlcgen.Queries
}

func NewCourseRepo(q *sqlcgen.Queries) *CourseRepo {
	return &CourseRepo{q: q}
}

func (r *CourseRepo) List(ctx context.Context) ([]entity.Course, error) {
	rows, err := r.q.ListCourses(ctx)
	if err != nil {
		return nil, err
	}
	courses := make([]entity.Course, len(rows))
	for i, row := range rows {
		courses[i] = entity.Course{
			ID:      row.ID,
			Name:    row.Name.String,
			Street:  row.Street.String,
			City:    row.City.String,
			State:   row.State.String,
			ZipCode: row.ZipCode.String,
			Phone:   row.Phone.String,
			Cost:    row.Cost.String,
		}
	}
	return courses, nil
}

func (r *CourseRepo) GetByNameAndCity(ctx context.Context, name, city string) (*entity.Course, error) {
	row, err := r.q.GetCourseByNameAndCity(ctx, sqlcgen.GetCourseByNameAndCityParams{
		Name: sql.NullString{String: name, Valid: true},
		City: sql.NullString{String: city, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.Course{
		ID:      row.ID,
		Name:    row.Name.String,
		Street:  row.Street.String,
		City:    row.City.String,
		State:   row.State.String,
		ZipCode: row.ZipCode.String,
		Phone:   row.Phone.String,
		Cost:    row.Cost.String,
	}, nil
}

func (r *CourseRepo) Create(ctx context.Context, c entity.Course) (*entity.Course, error) {
	row, err := r.q.CreateCourse(ctx, sqlcgen.CreateCourseParams{
		Name:    sql.NullString{String: c.Name, Valid: true},
		Street:  sql.NullString{String: c.Street, Valid: true},
		City:    sql.NullString{String: c.City, Valid: true},
		State:   sql.NullString{String: c.State, Valid: true},
		ZipCode: sql.NullString{String: c.ZipCode, Valid: true},
		Phone:   sql.NullString{String: c.Phone, Valid: true},
		Cost:    sql.NullString{String: c.Cost, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &entity.Course{
		ID:      row.ID,
		Name:    row.Name.String,
		Street:  row.Street.String,
		City:    row.City.String,
		State:   row.State.String,
		ZipCode: row.ZipCode.String,
		Phone:   row.Phone.String,
		Cost:    row.Cost.String,
	}, nil
}
