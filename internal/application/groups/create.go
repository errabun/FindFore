package groups

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func (s *Service) Create(ctx context.Context, in port.CreateGroupInput) (*port.GroupDetails, error) {
	if in.ActorID <= 0 {
		return nil, ErrInvalidGroup
	}
	if _, err := s.players.GetByID(ctx, in.ActorID); err != nil {
		return nil, ErrInvalidGroup
	}
	name, privacy, err := validateNamePrivacy(in.Name, in.Privacy)
	if err != nil {
		return nil, err
	}
	desc, err := clampDescription(in.Description)
	if err != nil {
		return nil, err
	}

	created, err := s.groups.CreateWithOwner(ctx, entity.Group{
		OwnerPlayerID: in.ActorID,
		Name:          name,
		Description:   desc,
		Privacy:       privacy,
	})
	if err != nil {
		return nil, err
	}
	viewer := &entity.GroupMembership{
		GroupID:  created.ID,
		PlayerID: in.ActorID,
		Role:     entity.GroupRoleOwner,
		Status:   entity.GroupMembershipActive,
	}
	return s.details(ctx, created, viewer)
}

func (s *Service) Update(ctx context.Context, in port.UpdateGroupInput) (*port.GroupDetails, error) {
	if in.ActorID <= 0 || in.GroupID <= 0 {
		return nil, ErrInvalidGroup
	}
	g, err := s.loadGroup(ctx, in.GroupID)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireManager(ctx, g.ID, in.ActorID); err != nil {
		return nil, err
	}
	name, privacy, err := validateNamePrivacy(in.Name, in.Privacy)
	if err != nil {
		return nil, err
	}
	desc, err := clampDescription(in.Description)
	if err != nil {
		return nil, err
	}
	g.Name = name
	g.Description = desc
	g.Privacy = privacy
	updated, err := s.groups.Update(ctx, *g)
	if err != nil {
		return nil, err
	}
	m, err := s.membership(ctx, updated.ID, in.ActorID)
	if err != nil {
		return nil, err
	}
	return s.details(ctx, updated, m)
}

func (s *Service) Get(ctx context.Context, actorID, groupID int64) (*port.GroupDetails, error) {
	if actorID <= 0 || groupID <= 0 {
		return nil, ErrGroupNotFound
	}
	g, err := s.loadGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	ok, viewer, err := s.canView(ctx, g, actorID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrGroupNotFound
	}
	return s.details(ctx, g, viewer)
}

func (s *Service) ListMine(ctx context.Context, actorID int64, limit, offset int32) ([]port.GroupDetails, error) {
	if actorID <= 0 {
		return nil, ErrInvalidGroup
	}
	list, err := s.groups.ListByPlayer(ctx, actorID, clampLimit(limit), clampOffset(offset))
	if err != nil {
		return nil, err
	}
	return s.detailsList(ctx, actorID, list)
}

func (s *Service) ListDiscover(ctx context.Context, actorID int64, search string, limit, offset int32) ([]port.GroupDetails, error) {
	if actorID <= 0 {
		return nil, ErrInvalidGroup
	}
	list, err := s.groups.ListPublic(ctx, search, clampLimit(limit), clampOffset(offset))
	if err != nil {
		return nil, err
	}
	return s.detailsList(ctx, actorID, list)
}

func validateIDs(ids ...int64) error {
	for _, id := range ids {
		if id <= 0 {
			return &apperr.ValidationError{Message: "invalid id"}
		}
	}
	return nil
}
