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
	providers  map[string]int64 // provider|externalID -> courseID
	nextID     int64
}

func newFakeCourseRepo() *fakeCourseRepo {
	return &fakeCourseRepo{
		byID:       make(map[int64]*entity.Course),
		byNameCity: make(map[string]*entity.Course),
		byProvider: make(map[string]*entity.Course),
		providers:  make(map[string]int64),
		nextID:     1,
	}
}

func nameCityKey(name, city string) string { return name + "|" + city }
func providerKey(provider, externalID string) string {
	return provider + "|" + externalID
}

func (r *fakeCourseRepo) List(context.Context) ([]entity.Course, error) { return nil, nil }

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

func (r *fakeCourseRepo) UpsertProvider(_ context.Context, courseID int64, provider, externalID string) error {
	key := providerKey(provider, externalID)
	r.providers[key] = courseID
	if c, ok := r.byID[courseID]; ok {
		cp := *c
		r.byProvider[key] = &cp
	}
	return nil
}

type noopSearcher struct{}

func (noopSearcher) Search(context.Context, string) ([]entity.Course, error) {
	return nil, nil
}

func TestFindOrCreateByProviderExternalID(t *testing.T) {
	repo := newFakeCourseRepo()
	svc := courses.NewService(repo, noopSearcher{})
	ctx := context.Background()

	first, created, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "Castle Pines", City: "Castle Rock", State: "CO",
		Provider: entity.ProviderGolfCourseAPI, ExternalID: "12345",
	})
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, int64(1), first.ID)
	assert.Equal(t, int64(1), repo.providers[providerKey(entity.ProviderGolfCourseAPI, "12345")])

	second, created, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "Castle Pines Golf Club", City: "Castle Rock", State: "CO",
		Provider: entity.ProviderGolfCourseAPI, ExternalID: "12345",
	})
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
	})
	require.NoError(t, err)

	again, created, err := svc.FindOrCreate(ctx, entity.Course{
		Name: "City Park", City: "Denver",
		Provider: entity.ProviderGolfCourseAPI, ExternalID: "99",
	})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, existing.ID, again.ID)
	assert.Equal(t, existing.ID, repo.providers[providerKey(entity.ProviderGolfCourseAPI, "99")])
}
