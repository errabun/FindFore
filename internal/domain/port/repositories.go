package port

import (
	"context"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PlayerRepository interface {
	List(ctx context.Context) ([]entity.Player, error)
	GetByID(ctx context.Context, id int64) (*entity.Player, error)
	GetByEmail(ctx context.Context, email string) (*entity.Player, error)
	GetByUsername(ctx context.Context, username string) (*entity.Player, error)
	Create(ctx context.Context, p entity.Player) (*entity.Player, error)
	Update(ctx context.Context, p entity.Player) (*entity.Player, error)
	GetPasswordByID(ctx context.Context, id int64) (string, error)
	UpdatePassword(ctx context.Context, id int64, passwordDigest string) error
	GetTokenVersion(ctx context.Context, id int64) (int32, error)
	ListIDsExcept(ctx context.Context, excludeID int64) ([]int64, error)
}

type CourseRepository interface {
	List(ctx context.Context) ([]entity.Course, error)
	GetByID(ctx context.Context, id int64) (*entity.Course, error)
	GetByNameAndCity(ctx context.Context, name, city string) (*entity.Course, error)
	GetByProviderExternalID(ctx context.Context, provider, externalID string) (*entity.Course, error)
	Create(ctx context.Context, c entity.Course) (*entity.Course, error)
	GetProvider(ctx context.Context, provider, externalID string) (*entity.CourseProvider, error)
	// GetProviderByCourse returns the provider link for a course + provider token.
	GetProviderByCourse(ctx context.Context, courseID int64, provider string) (*entity.CourseProvider, error)
	// LinkProvider associates (provider, external_id) with courseID.
	// Idempotent when the link already points at the same course; returns
	// entity.ErrProviderCourseConflict when it points at a different course.
	LinkProvider(ctx context.Context, courseID int64, provider, externalID string) error
}

type EventRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Event, error)
	GetDetailsByID(ctx context.Context, id int64) (*entity.EventWithDetails, error)
	ListAllIDs(ctx context.Context) ([]int64, error)
	ListPublicIDs(ctx context.Context) ([]int64, error)
	ListIDsByPlayerID(ctx context.Context, playerID int64) ([]int64, error)
	ListFriendsAvailableIDs(ctx context.Context, followerID int32, playerID int64) ([]int64, error)
	ListIDsByGroupID(ctx context.Context, groupID int64) ([]int64, error)
	ListUpcomingByGroupID(ctx context.Context, groupID int64) ([]entity.EventWithDetails, error)
	ListJoinableGroupDetails(ctx context.Context, actorID int64) ([]entity.EventWithDetails, error)
	Create(ctx context.Context, e entity.Event) (int64, error)
	CreateWithInvites(ctx context.Context, e entity.Event, invitees []int64) (int64, error)
	Update(ctx context.Context, e entity.Event) error
	Delete(ctx context.Context, id int64) error
	DeletePast(ctx context.Context) error
}

type TeeTimeRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.TeeTime, error)
	ListByCourseAndWindow(ctx context.Context, courseID int64, from, to time.Time) ([]entity.TeeTime, error)
	GetByProviderExternalID(ctx context.Context, provider, externalID string) (*entity.TeeTime, error)
	Create(ctx context.Context, t entity.TeeTime) (*entity.TeeTime, error)
	UpdateCache(ctx context.Context, t entity.TeeTime) (*entity.TeeTime, error)
	UpdateStatus(ctx context.Context, id int64, status string) (*entity.TeeTime, error)
	GetProvider(ctx context.Context, provider, externalID string) (*entity.TeeTimeProvider, error)
	// GetProviderByTeeTime returns the provider link for a tee time + provider token.
	GetProviderByTeeTime(ctx context.Context, teeTimeID int64, provider string) (*entity.TeeTimeProvider, error)
	// LinkProvider associates (provider, external_id) with teeTimeID.
	// Idempotent when the link already points at the same tee time; returns
	// entity.ErrProviderTeeTimeConflict when it points at a different tee time.
	LinkProvider(ctx context.Context, teeTimeID int64, provider, externalID string) error
}

type ReservationRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Reservation, error)
	GetActiveByTeeTimeID(ctx context.Context, teeTimeID int64) (*entity.Reservation, error)
	GetByClientIdempotency(ctx context.Context, bookedByPlayerID int64, clientIdempotencyKey string) (*entity.Reservation, error)
	Create(ctx context.Context, r entity.Reservation, players []entity.ReservationPlayer) (*entity.Reservation, error)
	Update(ctx context.Context, r entity.Reservation) (*entity.Reservation, error)
	ListPlayers(ctx context.Context, reservationID int64) ([]entity.ReservationPlayer, error)
}

type PlayerEventRepository interface {
	Get(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error)
	Create(ctx context.Context, pe entity.PlayerEvent) (*entity.PlayerEvent, error)
	UpdateStatus(ctx context.Context, playerID, eventID int64, status entity.InviteStatus) (*entity.PlayerEvent, error)
	ListPlayerIDsByEventAndStatus(ctx context.Context, eventID int64, status entity.InviteStatus) ([]int64, error)
	ListByEventIDs(ctx context.Context, eventIDs []int64) ([]entity.PlayerEvent, error)
	CountAcceptedForEvent(ctx context.Context, eventID int64) (int64, error)
	ClosePendingForEvent(ctx context.Context, eventID int64) error
	ReopenClosedForEvent(ctx context.Context, eventID int64) error
	// JoinAccepted locks the event row, enforces capacity, and inserts an accepted membership atomically.
	JoinAccepted(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error)
	// AcceptInvite locks the event row, enforces capacity, and sets an existing membership to accepted.
	AcceptInvite(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error)
}

type FriendshipRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Friendship, error)
	Find(ctx context.Context, requesterID, addresseeID int32) (*entity.Friendship, error)
	FindBetween(ctx context.Context, playerA, playerB int32) (*entity.Friendship, error)
	Create(ctx context.Context, requesterID, addresseeID int32, status entity.FriendshipStatus) (*entity.Friendship, error)
	UpdateStatus(ctx context.Context, id int64, status entity.FriendshipStatus) (*entity.Friendship, error)
	DeleteByID(ctx context.Context, id int64) error
	ListAcceptedFriendIDs(ctx context.Context, playerID int32) ([]int64, error)
	ListIncomingPending(ctx context.Context, addresseeID int32) ([]entity.Friendship, error)
	ListOutgoingPending(ctx context.Context, requesterID int32) ([]entity.Friendship, error)
	ListAcceptedEventIDs(ctx context.Context, playerID int64) ([]int64, error)
}

type PostRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.PostWithDetails, error)
	List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error)
	ListByGroupID(ctx context.Context, groupID int64, limit, offset int32) ([]entity.PostWithDetails, error)
	Create(ctx context.Context, playerID int64, body string, groupID *int64) (int64, error)
	Delete(ctx context.Context, id, playerID int64) error
	DeleteByID(ctx context.Context, id int64) error
}

type ReactionRepository interface {
	ListByPostID(ctx context.Context, postID int64) ([]entity.Reaction, error)
	Find(ctx context.Context, postID, playerID int64, emoji string) (*entity.Reaction, error)
	Create(ctx context.Context, postID, playerID int64, emoji string) (*entity.Reaction, error)
	Delete(ctx context.Context, postID, playerID int64, emoji string) error
}

type ReplyRepository interface {
	ListByPostID(ctx context.Context, postID int64) ([]entity.Reply, error)
	Create(ctx context.Context, postID, playerID int64, body string) (*entity.Reply, error)
	Delete(ctx context.Context, id, playerID int64) error
}

type GroupMemberRow struct {
	Membership entity.GroupMembership
	PlayerName string
}

type GroupInvitationRow struct {
	Invitation  entity.GroupInvitation
	GroupName   string
	InviterName string
	InviteeName string
}

type GroupRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Group, error)
	CreateWithOwner(ctx context.Context, g entity.Group) (*entity.Group, error)
	Update(ctx context.Context, g entity.Group) (*entity.Group, error)
	Delete(ctx context.Context, id int64) error
	TransferOwnership(ctx context.Context, groupID, fromPlayerID, toPlayerID int64) error
	ListPublic(ctx context.Context, search string, limit, offset int32) ([]entity.Group, error)
	ListByPlayer(ctx context.Context, playerID int64, limit, offset int32) ([]entity.Group, error)
	ListPublicSummaries(ctx context.Context, playerID int64, search string, limit, offset int32) ([]GroupDetails, error)
	ListByPlayerSummaries(ctx context.Context, playerID int64, limit, offset int32) ([]GroupDetails, error)
	CountActiveMembers(ctx context.Context, groupID int64) (int64, error)

	GetMembership(ctx context.Context, groupID, playerID int64) (*entity.GroupMembership, error)
	ListActiveMembers(ctx context.Context, groupID int64, limit, offset int32) ([]GroupMemberRow, error)
	ListPendingMembers(ctx context.Context, groupID int64) ([]GroupMemberRow, error)
	InsertMembership(ctx context.Context, m entity.GroupMembership) (*entity.GroupMembership, error)
	UpdateMembership(ctx context.Context, m entity.GroupMembership) (*entity.GroupMembership, error)
	DeleteMembership(ctx context.Context, groupID, playerID int64) error

	GetInvitationByID(ctx context.Context, id int64) (*entity.GroupInvitation, error)
	GetOutstandingInvitation(ctx context.Context, groupID, inviteeID int64) (*entity.GroupInvitation, error)
	ListInvitationsByInvitee(ctx context.Context, inviteeID int64) ([]GroupInvitationRow, error)
	ListOutstandingInvitations(ctx context.Context, groupID int64) ([]GroupInvitationRow, error)
	InsertInvitation(ctx context.Context, inv entity.GroupInvitation) (*entity.GroupInvitation, error)
	MarkInvitationAccepted(ctx context.Context, id, inviteeID int64) (*entity.GroupInvitation, error)
	MarkInvitationDeclined(ctx context.Context, id int64) (*entity.GroupInvitation, error)

	// AcceptInvitation marks the invite accepted and upserts an active member row in one transaction.
	AcceptInvitation(ctx context.Context, invitationID, playerID int64) (*entity.GroupMembership, error)
}
