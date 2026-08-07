package courses

import (
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

// Service manages course list, search, and persistence.
type Service struct {
	courses  port.CourseRepository
	searcher port.GolfCourseSearcher
}

func NewService(courses port.CourseRepository, searcher port.GolfCourseSearcher) *Service {
	return &Service{courses: courses, searcher: searcher}
}
