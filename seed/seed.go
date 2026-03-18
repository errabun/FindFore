package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/store"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	q := store.New(db)

	// Clean existing data in correct order
	for _, table := range []string{"player_events", "friendships", "events", "courses", "players"} {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			log.Fatalf("Failed to clean %s: %v", table, err)
		}
	}

	// Reset sequences
	for _, table := range []string{"players", "courses", "events", "friendships", "player_events"} {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table)); err != nil {
			log.Fatalf("Failed to reset %s sequence: %v", table, err)
		}
	}

	passwordHash, err := auth.HashPassword("password")
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create players
	players := []store.CreatePlayerParams{
		{Name: ns("Amy"), Phone: ns("2533597214"), Email: ns("jaharamclark@gmail.com"), Username: ns("username1"), PasswordDigest: ns(passwordHash)},
		{Name: ns("Andrew"), Phone: ns("3197952720"), Email: ns("keegan.oshea9@gmail.com"), Username: ns("username2"), PasswordDigest: ns(passwordHash)},
		{Name: ns("Amber"), Phone: ns("9999991236"), Email: ns("test3@test.com"), Username: ns("username3"), PasswordDigest: ns(passwordHash)},
		{Name: ns("Betty"), Phone: ns("9999991237"), Email: ns("test4@test.com"), Username: ns("username4"), PasswordDigest: ns(passwordHash)},
		{Name: ns("Burt"), Phone: ns("9999991238"), Email: ns("test5@test.com"), Username: ns("username5"), PasswordDigest: ns(passwordHash)},
		{Name: ns("Cleo"), Phone: ns("9999991239"), Email: ns("test6@test.com"), Username: ns("username6"), PasswordDigest: ns(passwordHash)},
		{Name: ns("Eric Rabun"), Phone: ns("9999991240"), Email: ns("errabun@gmail.com"), Username: ns("errabun"), PasswordDigest: ns(passwordHash)},
	}

	for _, p := range players {
		_, err := q.CreatePlayer(ctx, p)
		if err != nil {
			log.Fatalf("Failed to create player %s: %v", p.Name.String, err)
		}
	}
	fmt.Println("Created 7 players")

	// Create courses (using raw SQL to match Rails IDs)
	courses := []struct {
		name, street, city, state, zip, phone, cost string
	}{
		{"Green Valley Ranch Golf Club", "4900 Himalaya Road", "Denver", "Colorado", "80249", "303.371.3131", "80"},
		{"City Park Golf Course", "3181 E. 23rd Avenue", "Denver", "Colorado", "80205", "720.865.3410", "65"},
		{"Riverdale Golf Club", "13300 Riverdale Road", "Brighton", "Colorado", "80602", "303.659.4700", "74"},
		{"Willis Case Golf Course", "4999 Vrain Street", "Denver", "Colorado", "80212", "720.865.0700", "58"},
	}

	for _, c := range courses {
		_, err := db.ExecContext(ctx,
			"INSERT INTO courses (name, street, city, state, zip_code, phone, cost, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())",
			c.name, c.street, c.city, c.state, c.zip, c.phone, c.cost)
		if err != nil {
			log.Fatalf("Failed to create course %s: %v", c.name, err)
		}
	}
	fmt.Println("Created 4 courses")

	// Create events
	events := []store.CreateEventParams{
		{CourseID: ni32(1), Date: ns("08-01-2021"), TeeTime: ns("13:20"), OpenSpots: ni32(3), NumberOfHoles: ns("9"), HostID: ni32(1), Private: nb(true)},
		{CourseID: ni32(2), Date: ns("08-05-2021"), TeeTime: ns("14:20"), OpenSpots: ni32(4), NumberOfHoles: ns("18"), HostID: ni32(2), Private: nb(true)},
		{CourseID: ni32(3), Date: ns("08-10-2021"), TeeTime: ns("15:20"), OpenSpots: ni32(2), NumberOfHoles: ns("9"), HostID: ni32(3), Private: nb(false)},
		{CourseID: ni32(4), Date: ns("09-30-2021"), TeeTime: ns("15:20"), OpenSpots: ni32(2), NumberOfHoles: ns("9"), HostID: ni32(4), Private: nb(false)},
	}

	for _, e := range events {
		_, err := q.CreateEvent(ctx, e)
		if err != nil {
			log.Fatalf("Failed to create event: %v", err)
		}
	}
	fmt.Println("Created 4 events")

	// Create friendships
	friendships := [][2]int32{
		{1, 2}, {1, 3}, {2, 1}, {2, 3}, {3, 1}, {3, 2},
		{4, 5}, {4, 6}, {5, 4}, {5, 6},
	}

	for _, f := range friendships {
		_, err := q.CreateFriendship(ctx, store.CreateFriendshipParams{
			FollowerID: sql.NullInt32{Int32: f[0], Valid: true},
			FolloweeID: sql.NullInt32{Int32: f[1], Valid: true},
		})
		if err != nil {
			log.Fatalf("Failed to create friendship %d->%d: %v", f[0], f[1], err)
		}
	}
	fmt.Println("Created 10 friendships")

	// Create player events
	playerEvents := []struct {
		playerID int64
		eventID  int64
		status   int32 // 0=pending, 1=accepted, 2=declined
	}{
		{1, 1, 1}, {2, 1, 0}, {3, 1, 0},
		{1, 2, 0}, {2, 2, 1}, {3, 2, 0}, {6, 2, 0}, {5, 2, 0},
		{3, 3, 1}, {1, 3, 0}, {2, 3, 0}, {4, 3, 0}, {5, 3, 0}, {6, 3, 0},
	}

	for _, pe := range playerEvents {
		_, err := q.CreatePlayerEvent(ctx, store.CreatePlayerEventParams{
			PlayerID:     sql.NullInt64{Int64: pe.playerID, Valid: true},
			EventID:      sql.NullInt64{Int64: pe.eventID, Valid: true},
			InviteStatus: sql.NullInt32{Int32: pe.status, Valid: true},
		})
		if err != nil {
			log.Fatalf("Failed to create player_event p%d/e%d: %v", pe.playerID, pe.eventID, err)
		}
	}
	fmt.Println("Created 14 player events")

	fmt.Println("Seed complete!")
}

func ns(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
func ni32(i int32) sql.NullInt32 { return sql.NullInt32{Int32: i, Valid: true} }
func nb(b bool) sql.NullBool     { return sql.NullBool{Bool: b, Valid: true} }
