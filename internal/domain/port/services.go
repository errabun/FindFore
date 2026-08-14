package port

import (
	"context"
	"time"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PlayerService interface {
	List(ctx context.Context) ([]entity.PlayerWithDetails, error)
	GetWithDetails(ctx context.Context, id int64) (*entity.PlayerWithDetails, error)
	Create(ctx context.Context, name, phone, email, username, password, passwordConfirmation string) (*entity.Player, error)
	Update(ctx context.Context, callerID int64, name, phone, email, username string) (*entity.PlayerWithDetails, error)
	ChangePassword(ctx context.Context, callerID int64, currentPassword, newPassword, passwordConfirmation string) error
}

type SessionService interface {
	Login(ctx context.Context, login, password string) (*entity.PlayerWithDetails, string, error)
}

type CourseService interface {
	List(ctx context.Context) ([]entity.Course, error)
	Search(ctx context.Context, query string) ([]entity.CourseSearchResult, error)
	FindOrCreate(ctx context.Context, c entity.Course, link *entity.CourseProvider) (*entity.Course, bool, error)
}

type EventService interface {
	List(ctx context.Context, actorID int64, forPlayerID *int64, publicOnly bool) ([]entity.EventWithDetails, error)
	Get(ctx context.Context, id, viewerID int64) (*entity.EventWithDetails, error)
	Create(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error)
	Update(ctx context.Context, actorID int64, e entity.Event, invitees []int64) (*entity.EventWithDetails, error)
	Delete(ctx context.Context, actorID, id int64) error
	ListFriendsEvents(ctx context.Context, actorID int64) ([]entity.EventWithDetails, error)
	ListForGroup(ctx context.Context, actorID, groupID int64) ([]entity.EventWithDetails, error)
}

type PlayerEventService interface {
	UpdateStatus(ctx context.Context, playerID, eventID int64, status string) (*entity.PlayerEvent, error)
	JoinEvent(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error)
}

type FriendshipService interface {
	Request(ctx context.Context, actorID, playerID int32) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error)
	Accept(ctx context.Context, actorID int32, friendshipID int64) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error)
	Decline(ctx context.Context, actorID int32, friendshipID int64) error
	CancelOrUnfriend(ctx context.Context, actorID int32, friendshipID int64) error
	ListIncomingRequests(ctx context.Context, actorID int32) ([]entity.Friendship, error)
	ListOutgoingPendingIDs(ctx context.Context, actorID int32) ([]int64, error)
	ListAccepted(ctx context.Context, actorID int32) ([]entity.Friendship, error)
}

type PostService interface {
	List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error)
	ListForGroup(ctx context.Context, actorID, groupID int64, limit, offset int32) ([]entity.PostWithDetails, error)
	Create(ctx context.Context, playerID int64, body string) (*entity.PostWithDetails, error)
	CreateForGroup(ctx context.Context, actorID, groupID int64, body string) (*entity.PostWithDetails, error)
	Delete(ctx context.Context, postID, playerID int64) error
	ToggleReaction(ctx context.Context, postID, playerID int64, emoji string) ([]entity.Reaction, error)
	CreateReply(ctx context.Context, postID, playerID int64, body string) (*entity.Reply, error)
	DeleteReply(ctx context.Context, replyID, playerID int64) error
}

type GolfCourseSearcher interface {
	Search(ctx context.Context, query string) ([]entity.CourseSearchResult, error)
}

// BeginBookingInput is the FindFore-ID-only input for starting a reservation.
type BeginBookingInput struct {
	ActorID              int64
	TeeTimeID            int64
	Players              []entity.ReservationPlayer
	ClientIdempotencyKey string // Required; HTTP Idempotency-Key (not provider_request_id)
}

// BeginBookingResult reports the reservation and whether a new row was created.
type BeginBookingResult struct {
	Reservation *entity.Reservation
	Created     bool
}

// AvailabilitySource identifies where SearchAvailabilityResult.TeeTimes came from.
const (
	AvailabilitySourceProvider = "provider"
	AvailabilitySourceCache    = "cache"
)

// SearchAvailabilityResult is a best-effort tee-time list with freshness metadata.
// Cached rows are never a booking guarantee; ?players= is a filter on available_slots only.
type SearchAvailabilityResult struct {
	TeeTimes  []entity.TeeTime
	Source    string    // AvailabilitySourceProvider | AvailabilitySourceCache
	FetchedAt time.Time // UTC; provider success = now; cache fallback = newest LastSyncedAt when known
}

// BookingService is the provider-agnostic booking application API (HTTP boundary).
type BookingService interface {
	SearchAvailability(ctx context.Context, courseID int64, from, to time.Time, minPlayers int32) (*SearchAvailabilityResult, error)
	BeginBooking(ctx context.Context, in BeginBookingInput) (*BeginBookingResult, error)
	ConfirmBooking(ctx context.Context, actorID, reservationID int64) (*entity.Reservation, error)
	CancelBooking(ctx context.Context, actorID, reservationID int64) (*entity.Reservation, error)
	ListReservationPlayers(ctx context.Context, reservationID int64) ([]entity.ReservationPlayer, error)
}

type GroupDetails struct {
	Group       entity.Group
	OwnerName   string
	MemberCount int64
	Viewer      *entity.GroupMembership
}

type GroupMember struct {
	PlayerID   int64
	PlayerName string
	Role       string
	Status     string
}

type GroupInvitationView struct {
	Invitation  entity.GroupInvitation
	GroupName   string
	InviterName string
	InviteeName string
}

type CreateGroupInput struct {
	ActorID     int64
	Name        string
	Description string
	Privacy     string
}

type UpdateGroupInput struct {
	ActorID     int64
	GroupID     int64
	Name        string
	Description string
	Privacy     string
}

type GroupService interface {
	Create(ctx context.Context, in CreateGroupInput) (*GroupDetails, error)
	Get(ctx context.Context, actorID, groupID int64) (*GroupDetails, error)
	ListMine(ctx context.Context, actorID int64, limit, offset int32) ([]GroupDetails, error)
	ListDiscover(ctx context.Context, actorID int64, search string, limit, offset int32) ([]GroupDetails, error)
	Update(ctx context.Context, in UpdateGroupInput) (*GroupDetails, error)
	Join(ctx context.Context, actorID, groupID int64) (*entity.GroupMembership, error)
	Leave(ctx context.Context, actorID, groupID int64) error
	ListMembers(ctx context.Context, actorID, groupID int64, limit, offset int32) ([]GroupMember, error)
	RemoveMember(ctx context.Context, actorID, groupID, playerID int64) error
	Invite(ctx context.Context, actorID, groupID, inviteeID int64) (*entity.GroupInvitation, error)
	ListMyInvitations(ctx context.Context, actorID int64) ([]GroupInvitationView, error)
	ListGroupInvitations(ctx context.Context, actorID, groupID int64) ([]GroupInvitationView, error)
	CancelInvitation(ctx context.Context, actorID, groupID, invitationID int64) error
	AcceptInvitation(ctx context.Context, actorID, invitationID int64) (*entity.GroupMembership, error)
	DeclineInvitation(ctx context.Context, actorID, invitationID int64) error
	TransferOwnership(ctx context.Context, actorID, groupID, newOwnerID int64) (*GroupDetails, error)
	Delete(ctx context.Context, actorID, groupID int64) error
	ListJoinRequests(ctx context.Context, actorID, groupID int64) ([]GroupMember, error)
	ApproveJoinRequest(ctx context.Context, actorID, groupID, playerID int64) (*entity.GroupMembership, error)
	DenyJoinRequest(ctx context.Context, actorID, groupID, playerID int64) error
}
