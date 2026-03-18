package service

import (
	"context"
	"fmt"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type CourseService struct {
	courses  port.CourseRepository
	searcher port.GolfCourseSearcher
}

func NewCourseService(courses port.CourseRepository, searcher port.GolfCourseSearcher) *CourseService {
	return &CourseService{courses: courses, searcher: searcher}
}

func (s *CourseService) List(ctx context.Context) ([]entity.Course, error) {
	return s.courses.List(ctx)
}

func (s *CourseService) Search(ctx context.Context, query string) ([]entity.Course, error) {
	return s.searcher.Search(ctx, query)
}

func (s *CourseService) FindOrCreate(ctx context.Context, c entity.Course) (*entity.Course, error) {
	existing, err := s.courses.GetByNameAndCity(ctx, c.Name, c.City)
	if err == nil {
		return existing, nil
	}

	created, err := s.courses.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("create course: %w", err)
	}
	return created, nil
}
