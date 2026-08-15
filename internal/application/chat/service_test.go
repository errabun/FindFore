package chat

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type fakeGroups struct {
	details *port.GroupDetails
	members []port.GroupMember
	getErr  error
}

func (f *fakeGroups) Get(context.Context, int64, int64) (*port.GroupDetails, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.details, nil
}
func (f *fakeGroups) ListMembers(context.Context, int64, int64, int32, int32) ([]port.GroupMember, error) {
	return f.members, nil
}
func (f *fakeGroups) Create(context.Context, port.CreateGroupInput) (*port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeGroups) ListMine(context.Context, int64, int32, int32) ([]port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeGroups) ListDiscover(context.Context, int64, string, int32, int32) ([]port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeGroups) Update(context.Context, port.UpdateGroupInput) (*port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeGroups) Join(context.Context, int64, int64) (*entity.GroupMembership, error) {
	return nil, nil
}
func (f *fakeGroups) Leave(context.Context, int64, int64) error { return nil }
func (f *fakeGroups) RemoveMember(context.Context, int64, int64, int64) error {
	return nil
}
func (f *fakeGroups) Invite(context.Context, int64, int64, int64) (*entity.GroupInvitation, error) {
	return nil, nil
}
func (f *fakeGroups) ListMyInvitations(context.Context, int64) ([]port.GroupInvitationView, error) {
	return nil, nil
}
func (f *fakeGroups) ListGroupInvitations(context.Context, int64, int64) ([]port.GroupInvitationView, error) {
	return nil, nil
}
func (f *fakeGroups) CancelInvitation(context.Context, int64, int64, int64) error {
	return nil
}
func (f *fakeGroups) AcceptInvitation(context.Context, int64, int64) (*entity.GroupMembership, error) {
	return nil, nil
}
func (f *fakeGroups) DeclineInvitation(context.Context, int64, int64) error { return nil }
func (f *fakeGroups) TransferOwnership(context.Context, int64, int64, int64) (*port.GroupDetails, error) {
	return nil, nil
}
func (f *fakeGroups) Delete(context.Context, int64, int64) error { return nil }
func (f *fakeGroups) ListJoinRequests(context.Context, int64, int64) ([]port.GroupMember, error) {
	return nil, nil
}
func (f *fakeGroups) ApproveJoinRequest(context.Context, int64, int64, int64) (*entity.GroupMembership, error) {
	return nil, nil
}
func (f *fakeGroups) DenyJoinRequest(context.Context, int64, int64, int64) error {
	return nil
}

type fakeProvider struct {
	key     string
	ensured []int64
}

func (f *fakeProvider) APIKey() string { return f.key }
func (f *fakeProvider) IssueToken(context.Context, int64, string) (string, error) {
	return "tok", nil
}
func (f *fakeProvider) EnsureGroupChannel(_ context.Context, groupID int64, _ string, _ []port.GroupMember) error {
	f.ensured = append(f.ensured, groupID)
	return nil
}

func TestGroupSessionActiveMember(t *testing.T) {
	provider := &fakeProvider{key: "pk"}
	svc := NewService(&fakeGroups{
		details: &port.GroupDetails{
			Group:  entity.Group{ID: 10, Name: "Saturday Morning Golf"},
			Viewer: &entity.GroupMembership{Status: entity.GroupMembershipActive, Role: entity.GroupRoleMember},
		},
		members: []port.GroupMember{{PlayerID: 1, PlayerName: "Eric", Status: entity.GroupMembershipActive}},
	}, provider)

	sess, err := svc.GroupSession(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Equal(t, "pk", sess.APIKey)
	require.Equal(t, "tok", sess.Token)
	require.Equal(t, "messaging", sess.ChannelType)
	require.Equal(t, "group_10", sess.ChannelID)
	require.Equal(t, "1", sess.UserID)
	require.Equal(t, "Eric", sess.UserName)
	require.Equal(t, []int64{10}, provider.ensured)
}

func TestGroupSessionPendingIsNotFound(t *testing.T) {
	svc := NewService(&fakeGroups{
		details: &port.GroupDetails{
			Group:  entity.Group{ID: 10, Name: "Private"},
			Viewer: &entity.GroupMembership{Status: entity.GroupMembershipPending, Role: entity.GroupRoleMember},
		},
	}, &fakeProvider{key: "pk"})

	_, err := svc.GroupSession(context.Background(), 2, 10)
	require.ErrorIs(t, err, groups.ErrGroupNotFound)
}

func TestGroupSessionStrangerIsNotFound(t *testing.T) {
	svc := NewService(&fakeGroups{
		getErr: groups.ErrGroupNotFound,
	}, &fakeProvider{key: "pk"})

	_, err := svc.GroupSession(context.Background(), 9, 10)
	require.ErrorIs(t, err, groups.ErrGroupNotFound)
}

func TestGroupSessionDisabled(t *testing.T) {
	svc := NewService(&fakeGroups{}, nil)
	_, err := svc.GroupSession(context.Background(), 1, 10)
	require.ErrorIs(t, err, ErrChatDisabled)
}
