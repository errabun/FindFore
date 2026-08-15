package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/golfcourseapi"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/lightspeed"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/streamchat"
	"github.com/ericrabun/findfore-go/internal/application/booking"
	"github.com/ericrabun/findfore-go/internal/application/chat"
	"github.com/ericrabun/findfore-go/internal/application/courses"
	"github.com/ericrabun/findfore-go/internal/application/events"
	"github.com/ericrabun/findfore-go/internal/application/feed"
	"github.com/ericrabun/findfore-go/internal/application/friends"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/application/players"
	"github.com/ericrabun/findfore-go/internal/application/sessions"
	"github.com/ericrabun/findfore-go/internal/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	queries := sqlcgen.New(db)

	playerRepo := postgres.NewPlayerRepo(queries)
	courseRepo := postgres.NewCourseRepo(queries)
	eventRepo := postgres.NewEventRepo(queries, db)
	playerEventRepo := postgres.NewPlayerEventRepo(queries, db)
	friendshipRepo := postgres.NewFriendshipRepo(queries)
	postRepo := postgres.NewPostRepo(queries)
	reactionRepo := postgres.NewReactionRepo(queries)
	replyRepo := postgres.NewReplyRepo(queries)

	golfCourseClient := golfcourseapi.NewClient(os.Getenv("GOLF_COURSE_API_KEY"))
	teeTimeRepo := postgres.NewTeeTimeRepo(queries)
	reservationRepo := postgres.NewReservationRepo(queries, db)
	bookingProvider := lightspeed.NewClient(os.Getenv("LIGHTSPEED_BASE_URL"), os.Getenv("LIGHTSPEED_API_KEY"))
	bookingSvc := booking.NewService(teeTimeRepo, reservationRepo, courseRepo, playerRepo, bookingProvider)

	playerSvc := players.NewService(playerRepo, friendshipRepo)
	sessionSvc := sessions.NewService(playerRepo, friendshipRepo, cfg.JWTSecret)
	courseSvc := courses.NewService(courseRepo, golfCourseClient)
	friendshipSvc := friends.NewService(friendshipRepo, playerSvc)
	groupRepo := postgres.NewGroupRepo(queries, db)
	eventSvc := events.NewService(eventRepo, playerEventRepo, courseRepo, groupRepo)
	playerEventSvc := events.NewPlayerEventService(playerEventRepo, eventRepo, groupRepo)
	postSvc := feed.NewService(postRepo, reactionRepo, replyRepo, groupRepo)
	groupSvc := groups.NewService(groupRepo, playerRepo)

	h := httphandler.New(playerSvc, sessionSvc, courseSvc, eventSvc, playerEventSvc, friendshipSvc, postSvc, bookingSvc, groupSvc)
	if key, secret := os.Getenv("STREAM_API_KEY"), os.Getenv("STREAM_API_SECRET"); key != "" && secret != "" {
		adapter, err := streamchat.New(key, secret)
		if err != nil {
			slog.Error("failed to init stream chat", "err", err)
			os.Exit(1)
		}
		h = h.WithChat(chat.NewService(groupSvc, adapter))
		slog.Info("stream chat enabled")
	} else {
		slog.Info("stream chat disabled")
	}
	r := httphandler.NewRouter(h, cfg.JWTSecret, playerRepo)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
