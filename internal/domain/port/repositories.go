package port

import (
	"context"

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
	GetByNameAndCity(ctx context.Context, name, city string) (*entity.Course, error)
	Create(ctx context.Context, c entity.Course) (*entity.Course, error)
}

type EventRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.Event, error)
	GetDetailsByID(ctx context.Context, id int64) (*entity.EventWithDetails, error)
	ListAllIDs(ctx context.Context) ([]int64, error)
	ListPublicIDs(ctx context.Context) ([]int64, error)
	ListIDsByPlayerID(ctx context.Context, playerID int64) ([]int64, error)
	ListFriendsAvailableIDs(ctx context.Context, followerID int32, playerID int64) ([]int64, error)
	Create(ctx context.Context, e entity.Event) (int64, error)
	CreateWithInvites(ctx context.Context, e entity.Event, invitees []int64) (int64, error)
	Update(ctx context.Context, e entity.Event) error
	Delete(ctx context.Context, id int64) error
	DeletePast(ctx context.Context, today string) error
}

type PlayerEventRepository interface {
	Get(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error)
	Create(ctx context.Context, pe entity.PlayerEvent) (*entity.PlayerEvent, error)
	UpdateStatus(ctx context.Context, playerID, eventID int64, status entity.InviteStatus) (*entity.PlayerEvent, error)
	ListPlayerIDsByEventAndStatus(ctx context.Context, eventID int64, status entity.InviteStatus) ([]int64, error)
	CountAcceptedForEvent(ctx context.Context, eventID int64) (int64, error)
	ClosePendingForEvent(ctx context.Context, eventID int64) error
	ReopenClosedForEvent(ctx context.Context, eventID int64) error
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
	Create(ctx context.Context, playerID int64, body string) (int64, error)
	Delete(ctx context.Context, id, playerID int64) error
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
