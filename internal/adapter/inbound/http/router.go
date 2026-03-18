package httphandler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
)

func NewRouter(h *Handler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(mw.CorsHandler()))
	r.Use(mw.AuthOptional(jwtSecret))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/courses", h.ListCourses)
		r.Get("/courses/search", h.SearchCourses)
		r.Post("/courses", h.FindOrCreateCourse)

		r.Get("/players", h.ListPlayers)
		r.Post("/players", h.CreatePlayer)
		r.Get("/players/{player_id}/events", h.ListEvents)
		r.Get("/players/{player_id}/friends-events", h.ListFriendsEvents)

		r.Get("/events", h.ListEvents)
		r.Get("/event/{id}", h.GetEvent)
		r.Post("/event", h.CreateEvent)
		r.Patch("/event/{id}", h.UpdateEvent)
		r.Delete("/event/{id}", h.DeleteEvent)

		r.Post("/friendship", h.CreateFriendship)
		r.Delete("/friendship", h.DeleteFriendship)

		r.Patch("/player-event", h.UpdatePlayerEvent)
		r.Post("/player-event/join", h.JoinEvent)

		r.Get("/posts", h.ListPosts)
		r.Post("/posts", h.CreatePost)
		r.Delete("/posts/{post_id}", h.DeletePost)
		r.Post("/posts/{post_id}/reactions", h.ToggleReaction)
		r.Post("/posts/{post_id}/replies", h.CreateReply)
		r.Delete("/posts/{post_id}/replies/{reply_id}", h.DeleteReply)

		r.Post("/sessions", h.CreateSession)
	})

	// Serve frontend static files
	staticDir := frontendDistDir()
	if staticDir != "" {
		fileServer := http.FileServer(http.Dir(staticDir))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			// Serve static file if it exists (JS, CSS, images, etc.)
			path := filepath.Join(staticDir, r.URL.Path)
			if _, err := os.Stat(path); err == nil && !strings.HasSuffix(r.URL.Path, "/") {
				fileServer.ServeHTTP(w, r)
				return
			}
			// Otherwise serve index.html for SPA client-side routing
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		})
	}

	return r
}

func frontendDistDir() string {
	candidates := []string{"frontend/dist", "../frontend/dist"}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}
