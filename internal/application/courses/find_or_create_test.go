package courses_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/courses"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type fakeCourseRepo struct {
	byID       map[int64]*entity.Course
	byNameCity map[string]*entity.Course
	byProvider map[string]*entity.Course
	providers  map[string]*entity.CourseProvider // provider|externalID -> link
	nextID     int64
}

func newFakeCourseRepo() *fakeCourseRepo {
	return &fakeCourseRepo{
		byID:       make(map[int64]*entity.Course),
		byNameCity: make(map[string]*entity.Course),
		byProvider: make(map[string]*entity.Course),
		providers:  make(map[string]*entity.CourseProvider),
		nextID:     1,
	}
}

func nameCityKey(name, city string) string { return name + "|" + city }
func providerKey(provider, externalID string) string {
	return provider + "|" + externalID
}

func (r *fakeCourseRepo) List(context.Context) ([]entity.Course, error) { return nil, nil }

func (r *fakeCourseRepo) GetByID(_ context.Context, id int64) (*entity.Course, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCourseRepo) GetByNameAndCity(_ context.Context, name, city string) (*entity.Course, error) {
	c, ok := r.byNameCity[nameCityKey(name, city)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCourseRepo) GetByProviderExternalID(_ context.Context, provider, externalID string) (*entity.Course, error) {
	c, ok := r.byProvider[providerKey(provider, externalID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCourseRepo) Create(_ context.Context, c entity.Course) (*entity.Course, error) {
	c.ID = r.nextID
	r.nextID++
	cp := c
	r.byID[c.ID] = &cp
	r.byNameCity[nameCityKey(c.Name, c.City)] = &cp
	return &cp, nil
}

func (r *fakeCourseRepo) GetProvider(_ context.Context, provider, externalID string) (*entity.CourseProvider, error) {
	p, ok := r.providers[providerKey(provider, externalID)]
	if !ok {
		return nil, sql.ErrNoRows
	}
	cp := *p
	return &cp, nil
}

func (r *fakeCourseRepo) GetProviderByCourse(_ context.Context, courseID int64, provider string) (*entity.CourseProvider, error) {
	for _, p := range r.providers {
		if p.CourseID == courseID && p.Provider == provider {
			cp := *p
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeCourseRepo) LinkProvider(_ context.Context, courseID int64, provider, externalID string) error {
	key := providerKey(provider, externalID)
	if existing, ok := r.providers[key]; ok {
		if existing.CourseID == courseID {
			return nil
		}
		return entity.ErrProviderCourseConflict
	}
	link := &entity.CourseProvider{CourseID: courseID, Provider: provider, ExternalID: externalID}
	r.providers[key] = link
	if c, ok := r.byID[courseID]; ok {
		cp := *c
		r.byProvider[key] = &cp
	}
	return nil
}

type noopSearcher struct{}

func (noopSearcher) Search(context.Context, string) ([]entity.CourseSearchResult, error) {
	return nil, nil
}

func TestFindOrCreateByProviderExternalID(t *testing.T) {
	repo := newFakeCourseRepo()
	svc := courses.NewService(repo, noopSearcher{})
	ctx := context.Background()

	link := &entity.CourseProvider{Provider: entity.ProviderGolfCourseAPI, ExternalID: "12345"}
	first, created, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "Castle Pines", City: "Castle Rock", State: "CO",
	}, link)
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, int64(1), first.ID)

	second, created, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "Castle Pines Golf Club", City: "Castle Rock", State: "CO",
	}, link)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, first.ID, second.ID)
	assert.Len(t, repo.byID, 1)
}

func TestFindOrCreateByNameCityLinksProvider(t *testing.T) {
	repo := newFakeCourseRepo()
	svc := courses.NewService(repo, noopSearcher{})
	ctx := context.Background()

	existing, _, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "City Park", City: "Denver",
	}, nil)
	require.NoError(t, err)

	again, created, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "City Park", City: "Denver",
	}, &entity.CourseProvider{Provider: entity.ProviderGolfCourseAPI, ExternalID: "99"})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, existing.ID, again.ID)
	assert.Equal(t, existing.ID, repo.providers[providerKey(entity.ProviderGolfCourseAPI, "99")].CourseID)
}

func TestLinkProviderConflictDifferentCourse(t *testing.T) {
	repo := newFakeCourseRepo()
	svc := courses.NewService(repo, noopSearcher{})
	ctx := context.Background()

	link := &entity.CourseProvider{Provider: entity.ProviderGolfCourseAPI, ExternalID: "same"}
	a, _, err := svc.FindOrCreate(ctx, entity.Course{Name: "A", City: "Denver"}, link)
	require.NoError(t, err)

	_, _, err = svc.FindOrCreate(ctx, entity.Course{Name: "B", City: "Boulder"}, link)
	require.NoError(t, err) // resolved by provider → returns A
	again, created, err := svc.FindOrCreate(ctx, entity.Course{Name: "A", City: "Denver"}, link)
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, a.ID, again.ID)

	// Direct link conflict: same provider id onto a different course
	b, _, err := svc.FindOrCreate(ctx, entity.Course{Name: "B", City: "Boulder"}, nil)
	require.NoError(t, err)
	err = repo.LinkProvider(ctx, b.ID, entity.ProviderGolfCourseAPI, "same")
	require.ErrorIs(t, err, entity.ErrProviderCourseConflict)
}

func TestLinkProviderIdempotentSameCourse(t *testing.T) {
	repo := newFakeCourseRepo()
	ctx := context.Background()
	c, _, err := courses.NewService(repo, noopSearcher{}).FindOrCreate(ctx, entity.Course{Name: "X", City: "Y"},
		&entity.CourseProvider{Provider: "golfcourseapi", ExternalID: "1"})
	require.NoError(t, err)
	require.NoError(t, repo.LinkProvider(ctx, c.ID, "golfcourseapi", "1"))
}
