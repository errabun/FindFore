package groups

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

func clampLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func clampOffset(offset int32) int32 {
	if offset < 0 {
		return 0
	}
	return offset
}

func validateNamePrivacy(name, privacy string) (string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxNameLen {
		return "", "", &apperr.ValidationError{Message: "name is required (max 100 characters)"}
	}
	switch privacy {
	case entity.GroupPrivacyPublic, entity.GroupPrivacyPrivate:
	default:
		return "", "", &apperr.ValidationError{Message: "privacy must be public or private"}
	}
	return name, privacy, nil
}

func clampDescription(desc string) (string, error) {
	desc = strings.TrimSpace(desc)
	if len(desc) > maxDescriptionLen {
		return "", &apperr.ValidationError{Message: "description must be at most 1000 characters"}
	}
	return desc, nil
}

func (s *Service) loadGroup(ctx context.Context, groupID int64) (*entity.Group, error) {
	g, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return g, nil
}

func (s *Service) membership(ctx context.Context, groupID, playerID int64) (*entity.GroupMembership, error) {
	m, err := s.groups.GetMembership(ctx, groupID, playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) outstandingInvite(ctx context.Context, groupID, playerID int64) (*entity.GroupInvitation, error) {
	inv, err := s.groups.GetOutstandingInvitation(ctx, groupID, playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if !inv.IsOutstanding(time.Now().UTC()) {
		return nil, nil
	}
	return inv, nil
}

func (s *Service) canView(ctx context.Context, g *entity.Group, actorID int64) (bool, *entity.GroupMembership, error) {
	if g.IsPublic() {
		m, err := s.membership(ctx, g.ID, actorID)
		return true, m, err
	}
	m, err := s.membership(ctx, g.ID, actorID)
	if err != nil {
		return false, nil, err
	}
	if m != nil {
		return true, m, nil
	}
	inv, err := s.outstandingInvite(ctx, g.ID, actorID)
	if err != nil {
		return false, nil, err
	}
	return inv != nil, nil, nil
}

func (s *Service) requireManager(ctx context.Context, groupID, actorID int64) (*entity.GroupMembership, error) {
	m, err := s.membership(ctx, groupID, actorID)
	if err != nil {
		return nil, err
	}
	if m == nil || !m.CanManage() {
		return nil, ErrGroupForbidden
	}
	return m, nil
}

func (s *Service) requireOwner(ctx context.Context, groupID, actorID int64) (*entity.GroupMembership, error) {
	m, err := s.membership(ctx, groupID, actorID)
	if err != nil {
		return nil, err
	}
	if m == nil || !m.IsOwner() {
		return nil, ErrGroupForbidden
	}
	return m, nil
}

func (s *Service) details(ctx context.Context, g *entity.Group, viewer *entity.GroupMembership) (*port.GroupDetails, error) {
	owner, err := s.players.GetByID(ctx, g.OwnerPlayerID)
	ownerName := ""
	if err == nil {
		ownerName = owner.Name
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	count, err := s.groups.CountActiveMembers(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	return &port.GroupDetails{
		Group:       *g,
		OwnerName:   ownerName,
		MemberCount: count,
		Viewer:      viewer,
	}, nil
}

func (s *Service) detailsList(ctx context.Context, actorID int64, list []entity.Group) ([]port.GroupDetails, error) {
	out := make([]port.GroupDetails, 0, len(list))
	for i := range list {
		g := list[i]
		m, err := s.membership(ctx, g.ID, actorID)
		if err != nil {
			return nil, err
		}
		d, err := s.details(ctx, &g, m)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, nil
}
