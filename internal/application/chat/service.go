package chat

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

var ErrChatDisabled = errors.New("chat is not configured")

const (
	channelType    = "messaging"
	memberPageSize = int32(50)
	maxMemberPages = 20
)

type Service struct {
	groups   port.GroupService
	provider port.ChatProvider
}

func NewService(groups port.GroupService, provider port.ChatProvider) *Service {
	return &Service{groups: groups, provider: provider}
}

func (s *Service) GroupSession(ctx context.Context, actorID, groupID int64) (*port.GroupChatSession, error) {
	if s == nil || s.provider == nil {
		return nil, ErrChatDisabled
	}
	details, err := s.groups.Get(ctx, actorID, groupID)
	if err != nil {
		return nil, err
	}
	if details.Viewer == nil || !details.Viewer.IsActive() {
		return nil, groups.ErrGroupNotFound
	}

	members, err := s.listAllMembers(ctx, actorID, groupID)
	if err != nil {
		return nil, err
	}
	if err := s.provider.EnsureGroupChannel(ctx, groupID, details.Group.Name, members); err != nil {
		return nil, fmt.Errorf("ensure group channel: %w", err)
	}

	var userName string
	for _, m := range members {
		if m.PlayerID == actorID {
			userName = m.PlayerName
			break
		}
	}
	token, err := s.provider.IssueToken(ctx, actorID, userName)
	if err != nil {
		return nil, fmt.Errorf("issue chat token: %w", err)
	}
	return &port.GroupChatSession{
		APIKey:      s.provider.APIKey(),
		Token:       token,
		ChannelType: channelType,
		ChannelID:   fmt.Sprintf("group_%d", groupID),
		UserID:      strconv.FormatInt(actorID, 10),
		UserName:    userName,
	}, nil
}

func (s *Service) listAllMembers(ctx context.Context, actorID, groupID int64) ([]port.GroupMember, error) {
	var all []port.GroupMember
	for offset := int32(0); ; offset += memberPageSize {
		page, err := s.groups.ListMembers(ctx, actorID, groupID, memberPageSize, offset)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if int32(len(page)) < memberPageSize {
			return all, nil
		}
		if offset/memberPageSize+1 >= maxMemberPages {
			return nil, fmt.Errorf("group %d has too many members to sync chat", groupID)
		}
	}
}
