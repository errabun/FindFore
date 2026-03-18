package port

import (
	"context"

	"github.com/ericrabun/findfore-go/internal/domain/entity"
)

type PlayerService interface {
	List(ctx context.Context) ([]entity.PlayerWithDetails, error)
	GetWithDetails(ctx context.Context, id int64) (*entity.PlayerWithDetails, error)
	Create(ctx context.Context, name, phone, email, username, password, passwordConfirmation string) (*entity.Player, error)
}

type SessionService interface {
	Login(ctx context.Context, login, password string) (*entity.PlayerWithDetails, string, error)
}

type CourseService interface {
	List(ctx context.Context) ([]entity.Course, error)
	Search(ctx context.Context, query string) ([]entity.Course, error)
	FindOrCreate(ctx context.Context, c entity.Course) (*entity.Course, error)
}

type EventService interface {
	List(ctx context.Context, playerID *int64, publicOnly bool) ([]entity.EventWithDetails, error)
	Get(ctx context.Context, id int64) (*entity.EventWithDetails, error)
	Create(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error)
	Update(ctx context.Context, e entity.Event, invitees []int64) (*entity.EventWithDetails, error)
	Delete(ctx context.Context, id int64) error
	ListFriendsEvents(ctx context.Context, playerID int64) ([]entity.EventWithDetails, error)
}

type PlayerEventService interface {
	UpdateStatus(ctx context.Context, playerID, eventID int64, status string) (*entity.PlayerEvent, error)
	JoinEvent(ctx context.Context, playerID, eventID int64) (*entity.PlayerEvent, error)
}

type FriendshipService interface {
	FindOrCreate(ctx context.Context, followerID, followeeID int32) (*entity.Friendship, *entity.PlayerWithDetails, *entity.PlayerWithDetails, error)
	Delete(ctx context.Context, followerID, followeeID int32) error
}

type PostService interface {
	List(ctx context.Context, limit, offset int32) ([]entity.PostWithDetails, error)
	Create(ctx context.Context, playerID int64, body string) (*entity.PostWithDetails, error)
	Delete(ctx context.Context, postID, playerID int64) error
	ToggleReaction(ctx context.Context, postID, playerID int64, emoji string) ([]entity.Reaction, error)
	CreateReply(ctx context.Context, postID, playerID int64, body string) (*entity.Reply, error)
	DeleteReply(ctx context.Context, replyID, playerID int64) error
}

type GolfCourseSearcher interface {
	Search(ctx context.Context, query string) ([]entity.Course, error)
}
