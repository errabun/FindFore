package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	httphandler "github.com/ericrabun/findfore-go/internal/adapter/inbound/http"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/golfcourseapi"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres"
	"github.com/ericrabun/findfore-go/internal/adapter/outbound/postgres/sqlcgen"
	"github.com/ericrabun/findfore-go/internal/application/service"
	"github.com/ericrabun/findfore-go/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// sqlc queries
	queries := sqlcgen.New(db)

	// Repository adapters
	playerRepo := postgres.NewPlayerRepo(queries)
	courseRepo := postgres.NewCourseRepo(queries)
	eventRepo := postgres.NewEventRepo(queries, db)
	playerEventRepo := postgres.NewPlayerEventRepo(queries)
	friendshipRepo := postgres.NewFriendshipRepo(queries)
	postRepo := postgres.NewPostRepo(queries)
	reactionRepo := postgres.NewReactionRepo(queries)
	replyRepo := postgres.NewReplyRepo(queries)

	// External adapters
	golfCourseClient := golfcourseapi.NewClient(os.Getenv("GOLF_COURSE_API_KEY"))

	// Application services
	playerSvc := service.NewPlayerService(playerRepo, friendshipRepo)
	sessionSvc := service.NewSessionService(playerRepo, friendshipRepo, cfg.JWTSecret)
	courseSvc := service.NewCourseService(courseRepo, golfCourseClient)
	eventSvc := service.NewEventService(eventRepo, playerEventRepo)
	playerEventSvc := service.NewPlayerEventService(playerEventRepo, eventRepo)
	friendshipSvc := service.NewFriendshipService(friendshipRepo, playerSvc)
	postSvc := service.NewPostService(postRepo, reactionRepo, replyRepo)

	// HTTP handler + router
	h := httphandler.New(playerSvc, sessionSvc, courseSvc, eventSvc, playerEventSvc, friendshipSvc, postSvc)
	r := httphandler.NewRouter(h, cfg.JWTSecret)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
