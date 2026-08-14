package httphandler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
)

func NewRouter(h *Handler, jwtSecret string, tokenVersions mw.TokenVersionLookup) *chi.Mux {
	r := chi.NewRouter()

	r.Use(mw.RequestID)
	r.Use(mw.Recoverer)
	r.Use(mw.AccessLog)
	r.Use(cors.Handler(mw.CorsHandler()))

	loginLimiter := mw.NewLoginRateLimiter(10, 15*time.Minute)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no JWT required)
		r.Group(func(r chi.Router) {
			r.Use(mw.AuthOptional(jwtSecret))
			r.With(loginLimiter.Middleware).Post("/sessions", h.CreateSession)
			r.Post("/players", h.CreatePlayer)
			r.Get("/courses", h.ListCourses)
			r.Get("/courses/search", h.SearchCourses)
		})

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(mw.AuthRequired(jwtSecret, tokenVersions))

			r.Post("/courses", h.FindOrCreateCourse)
			r.Get("/courses/{courseID}/tee-times", h.ListCourseTeeTimes)

			r.Post("/reservations", h.CreateReservation)
			r.Post("/reservations/{id}/confirm", h.ConfirmReservation)
			r.Post("/reservations/{id}/cancel", h.CancelReservation)

			r.Get("/players", h.ListPlayers)
			r.Patch("/players/{player_id}", h.UpdatePlayer)
			r.Patch("/players/{player_id}/password", h.ChangePassword)
			r.Get("/players/{player_id}/events", h.ListEvents)
			r.Get("/players/{player_id}/friends-events", h.ListFriendsEvents)

			r.Get("/events", h.ListEvents)
			r.Get("/event/{id}", h.GetEvent)
			r.Post("/event", h.CreateEvent)
			r.Patch("/event/{id}", h.UpdateEvent)
			r.Delete("/event/{id}", h.DeleteEvent)

			r.Get("/friendships", h.ListFriendships)
			r.Get("/friendships/requests", h.ListFriendshipRequests)
			r.Get("/friendships/outgoing", h.ListOutgoingFriendshipRequests)
			r.Post("/friendships", h.CreateFriendship)
			r.Post("/friendships/{id}/accept", h.AcceptFriendship)
			r.Post("/friendships/{id}/decline", h.DeclineFriendship)
			r.Delete("/friendships/{id}", h.DeleteFriendship)

			r.Patch("/player-event", h.UpdatePlayerEvent)
			r.Post("/player-event/join", h.JoinEvent)

			r.Get("/posts", h.ListPosts)
			r.Post("/posts", h.CreatePost)
			r.Delete("/posts/{post_id}", h.DeletePost)
			r.Post("/posts/{post_id}/reactions", h.ToggleReaction)
			r.Post("/posts/{post_id}/replies", h.CreateReply)
			r.Delete("/posts/{post_id}/replies/{reply_id}", h.DeleteReply)

			r.Get("/groups", h.ListGroups)
			r.Post("/groups", h.CreateGroup)
			r.Get("/groups/{id}", h.GetGroup)
			r.Patch("/groups/{id}", h.UpdateGroup)
			r.Delete("/groups/{id}", h.DeleteGroup)
			r.Post("/groups/{id}/join", h.JoinGroup)
			r.Post("/groups/{id}/leave", h.LeaveGroup)
			r.Post("/groups/{id}/transfer-ownership", h.TransferGroupOwnership)
			r.Get("/groups/{id}/members", h.ListGroupMembers)
			r.Delete("/groups/{id}/members/{playerId}", h.RemoveGroupMember)
			r.Post("/groups/{id}/invitations", h.InviteToGroup)
			r.Get("/groups/{id}/invitations", h.ListGroupInvitations)
			r.Delete("/groups/{id}/invitations/{invitationId}", h.CancelGroupInvitation)
			r.Get("/groups/{id}/join-requests", h.ListJoinRequests)
			r.Post("/groups/{id}/join-requests/{playerId}/approve", h.ApproveJoinRequest)
			r.Post("/groups/{id}/join-requests/{playerId}/deny", h.DenyJoinRequest)
			r.Get("/groups/{id}/posts", h.ListGroupPosts)
			r.Post("/groups/{id}/posts", h.CreateGroupPost)
			r.Get("/groups/{id}/events", h.ListGroupEvents)
			r.Get("/group-invitations", h.ListMyGroupInvitations)
			r.Post("/group-invitations/{id}/accept", h.AcceptGroupInvitation)
			r.Post("/group-invitations/{id}/decline", h.DeclineGroupInvitation)
		})
	})

	// Serve frontend static files
	staticDir := frontendDistDir()
	if staticDir != "" {
		fileServer := http.FileServer(http.Dir(staticDir))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			path := filepath.Join(staticDir, r.URL.Path)
			if _, err := os.Stat(path); err == nil && !strings.HasSuffix(r.URL.Path, "/") {
				fileServer.ServeHTTP(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
	}

	return r
}

func frontendDistDir() string {
	candidates := []string{"frontend/dist", "../frontend/dist", "/frontend/dist"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}
