package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ericrabun/findfore-go/internal/auth"
	"github.com/ericrabun/findfore-go/internal/model"
	"github.com/ericrabun/findfore-go/internal/store"
)

var (
	testDB      *sql.DB
	testQueries *store.Queries
	testHandler *Handler
)

const testJWTSecret = "test-jwt-secret"

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/fore-finder-be_go_test?sslmode=disable"
	} else {
		// Use test database
		dbURL = "postgres://localhost:5432/fore-finder-be_go_test?sslmode=disable"
	}

	var err error
	// Try to create the test database
	adminDB, err := sql.Open("postgres", "postgres://localhost:5432/postgres?sslmode=disable")
	if err == nil {
		adminDB.Exec("CREATE DATABASE \"fore-finder-be_go_test\"")
		adminDB.Close()
	}

	testDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Create tables
	createTables(testDB)

	testQueries = store.New(testDB)
	testHandler = New(testQueries, testDB, testJWTSecret)

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func createTables(db *sql.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR, phone VARCHAR, email VARCHAR, username VARCHAR, password_digest VARCHAR,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS courses (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR, street VARCHAR, city VARCHAR, state VARCHAR, zip_code VARCHAR, phone VARCHAR, cost VARCHAR,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS events (
		id BIGSERIAL PRIMARY KEY,
		course_id INTEGER, date VARCHAR, tee_time VARCHAR, open_spots INTEGER, number_of_holes VARCHAR,
		private BOOLEAN, host_id INTEGER,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS friendships (
		id BIGSERIAL PRIMARY KEY,
		follower_id INTEGER, followee_id INTEGER,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS player_events (
		id BIGSERIAL PRIMARY KEY,
		player_id BIGINT, event_id BIGINT, invite_status INTEGER DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_pe_event_id ON player_events (event_id);
	CREATE INDEX IF NOT EXISTS idx_pe_player_id ON player_events (player_id);
	`
	db.Exec(schema)

	// Add FK constraint if not exists (ignore error if already exists)
	db.Exec("ALTER TABLE player_events ADD CONSTRAINT fk_pe_events FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE")
	db.Exec("ALTER TABLE player_events ADD CONSTRAINT fk_pe_players FOREIGN KEY (player_id) REFERENCES players(id)")
}

func cleanDB(t *testing.T) {
	t.Helper()
	for _, table := range []string{"player_events", "friendships", "events", "courses", "players"} {
		testDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
	for _, table := range []string{"players", "courses", "events", "friendships", "player_events"} {
		testDB.Exec(fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", table))
	}
}

func seedPlayer(t *testing.T, name, email, password string) int64 {
	t.Helper()
	hash, _ := auth.HashPassword(password)
	row := testDB.QueryRow(
		"INSERT INTO players (name, phone, email, username, password_digest, created_at, updated_at) VALUES ($1, '5551234', $2, $3, $4, NOW(), NOW()) RETURNING id",
		name, email, name, hash,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		t.Fatalf("seedPlayer failed: %v", err)
	}
	return id
}

func seedCourse(t *testing.T, name string) int64 {
	t.Helper()
	row := testDB.QueryRow(
		"INSERT INTO courses (name, street, city, state, zip_code, phone, cost, created_at, updated_at) VALUES ($1, '123 Main St', 'Denver', 'CO', '80202', '555-1234', '80', NOW(), NOW()) RETURNING id",
		name,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		t.Fatalf("seedCourse failed: %v", err)
	}
	return id
}

func seedEvent(t *testing.T, courseID, hostID int64, openSpots int, private bool) int64 {
	t.Helper()
	row := testDB.QueryRow(
		"INSERT INTO events (course_id, date, tee_time, open_spots, number_of_holes, private, host_id, created_at, updated_at) VALUES ($1, '2025-01-01', '10:00', $2, '18', $3, $4, NOW(), NOW()) RETURNING id",
		courseID, openSpots, private, hostID,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		t.Fatalf("seedEvent failed: %v", err)
	}
	return id
}

func seedPlayerEvent(t *testing.T, playerID, eventID int64, status int) {
	t.Helper()
	_, err := testDB.Exec(
		"INSERT INTO player_events (player_id, event_id, invite_status, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
		playerID, eventID, status,
	)
	if err != nil {
		t.Fatalf("seedPlayerEvent failed: %v", err)
	}
}

func seedFriendship(t *testing.T, followerID, followeeID int64) {
	t.Helper()
	_, err := testDB.Exec(
		"INSERT INTO friendships (follower_id, followee_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())",
		followerID, followeeID,
	)
	if err != nil {
		t.Fatalf("seedFriendship failed: %v", err)
	}
}

func doRequest(t *testing.T, method, path string, body interface{}, handler http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func doRequestWithChiCtx(t *testing.T, method, path string, body interface{}, handler http.HandlerFunc, params map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// ===================== COURSES =====================

func TestListCourses(t *testing.T) {
	cleanDB(t)
	seedCourse(t, "Green Valley Ranch")
	seedCourse(t, "City Park")

	rr := doRequest(t, "GET", "/api/v1/courses", nil, testHandler.ListCourses)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var courses []model.CourseResponse
	json.NewDecoder(rr.Body).Decode(&courses)
	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}
	if courses[0].Name != "Green Valley Ranch" {
		t.Errorf("expected first course name 'Green Valley Ranch', got '%s'", courses[0].Name)
	}
}

// ===================== PLAYERS =====================

func TestListPlayers(t *testing.T) {
	cleanDB(t)
	seedPlayer(t, "Amy", "amy@test.com", "password")
	seedPlayer(t, "Bob", "bob@test.com", "password")

	rr := doRequest(t, "GET", "/api/v1/players", nil, testHandler.ListPlayers)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var players []model.PlayerResponse
	json.NewDecoder(rr.Body).Decode(&players)
	if len(players) != 2 {
		t.Errorf("expected 2 players, got %d", len(players))
	}
}

func TestCreatePlayer_Success(t *testing.T) {
	cleanDB(t)

	body := map[string]string{
		"name":                  "Test Player",
		"phone":                 "5551234",
		"email":                 "test@test.com",
		"username":              "testuser",
		"password":              "password",
		"password_confirmation": "password",
	}

	rr := doRequest(t, "POST", "/api/v1/players", body, testHandler.CreatePlayer)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var player model.PlayerResponse
	json.NewDecoder(rr.Body).Decode(&player)
	if player.Name != "Test Player" {
		t.Errorf("expected name 'Test Player', got '%s'", player.Name)
	}
	if player.Email != "test@test.com" {
		t.Errorf("expected email 'test@test.com', got '%s'", player.Email)
	}
}

func TestCreatePlayer_InvalidEmail(t *testing.T) {
	cleanDB(t)

	body := map[string]string{
		"name":                  "Test Player",
		"phone":                 "5551234",
		"email":                 "not-an-email",
		"username":              "testuser",
		"password":              "password",
		"password_confirmation": "password",
	}

	rr := doRequest(t, "POST", "/api/v1/players", body, testHandler.CreatePlayer)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestCreatePlayer_PasswordMismatch(t *testing.T) {
	cleanDB(t)

	body := map[string]string{
		"name":                  "Test Player",
		"phone":                 "5551234",
		"email":                 "test@test.com",
		"username":              "testuser",
		"password":              "password",
		"password_confirmation": "different",
	}

	rr := doRequest(t, "POST", "/api/v1/players", body, testHandler.CreatePlayer)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestCreatePlayer_MissingName(t *testing.T) {
	cleanDB(t)

	body := map[string]string{
		"phone":                 "5551234",
		"email":                 "test@test.com",
		"username":              "testuser",
		"password":              "password",
		"password_confirmation": "password",
	}

	rr := doRequest(t, "POST", "/api/v1/players", body, testHandler.CreatePlayer)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

// ===================== SESSIONS =====================

func TestCreateSession_Success(t *testing.T) {
	cleanDB(t)
	seedPlayer(t, "Amy", "amy@test.com", "password")

	body := map[string]string{
		"email":    "amy@test.com",
		"password": "password",
	}

	rr := doRequest(t, "POST", "/api/v1/sessions", body, testHandler.CreateSession)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var login model.LoginResponse
	json.NewDecoder(rr.Body).Decode(&login)
	if login.Token == "" {
		t.Error("expected non-empty token")
	}
	if login.Name != "Amy" {
		t.Errorf("expected name 'Amy', got '%s'", login.Name)
	}
}

func TestCreateSession_WrongPassword(t *testing.T) {
	cleanDB(t)
	seedPlayer(t, "Amy", "amy@test.com", "password")

	body := map[string]string{
		"email":    "amy@test.com",
		"password": "wrongpassword",
	}

	rr := doRequest(t, "POST", "/api/v1/sessions", body, testHandler.CreateSession)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestCreateSession_WrongEmail(t *testing.T) {
	cleanDB(t)
	seedPlayer(t, "Amy", "amy@test.com", "password")

	body := map[string]string{
		"email":    "wrong@test.com",
		"password": "password",
	}

	rr := doRequest(t, "POST", "/api/v1/sessions", body, testHandler.CreateSession)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

// ===================== EVENTS =====================

func TestListEvents(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	seedEvent(t, c1, p1, 3, false)
	seedEvent(t, c1, p1, 2, true)

	// Seed player_events so the player has events
	seedPlayerEvent(t, p1, 1, 1) // accepted
	seedPlayerEvent(t, p1, 2, 1) // accepted

	rr := doRequest(t, "GET", "/api/v1/events", nil, testHandler.ListEvents)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var events []model.EventResponse
	json.NewDecoder(rr.Body).Decode(&events)
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestListEvents_PublicOnly(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	seedEvent(t, c1, p1, 3, false)
	seedEvent(t, c1, p1, 2, true)

	req := httptest.NewRequest("GET", "/api/v1/events?private=false", nil)
	rr := httptest.NewRecorder()
	testHandler.ListEvents(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var events []model.EventResponse
	json.NewDecoder(rr.Body).Decode(&events)
	if len(events) != 1 {
		t.Errorf("expected 1 public event, got %d", len(events))
	}
}

func TestGetEvent(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, p1, 3, false)
	seedPlayerEvent(t, p1, eid, 1) // accepted

	rr := doRequestWithChiCtx(t, "GET", "/api/v1/event/1", nil, testHandler.GetEvent, map[string]string{"id": fmt.Sprintf("%d", eid)})

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var event model.EventResponse
	json.NewDecoder(rr.Body).Decode(&event)
	if event.CourseName != "Green Valley" {
		t.Errorf("expected course name 'Green Valley', got '%s'", event.CourseName)
	}
	if event.HostName != "Amy" {
		t.Errorf("expected host name 'Amy', got '%s'", event.HostName)
	}
	if len(event.Accepted) != 1 {
		t.Errorf("expected 1 accepted player, got %d", len(event.Accepted))
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	cleanDB(t)

	rr := doRequestWithChiCtx(t, "GET", "/api/v1/event/999", nil, testHandler.GetEvent, map[string]string{"id": "999"})

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestCreateEvent_Public(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	_ = p2
	c1 := seedCourse(t, "Green Valley")

	body := map[string]interface{}{
		"course_id":       c1,
		"date":            "2025-08-01",
		"tee_time":        "10:00",
		"open_spots":      3,
		"number_of_holes": "18",
		"private":         false,
		"host_id":         p1,
		"invitees":        []int64{},
	}

	rr := doRequest(t, "POST", "/api/v1/event", body, testHandler.CreateEvent)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var event model.EventResponse
	json.NewDecoder(rr.Body).Decode(&event)
	if event.CourseName != "Green Valley" {
		t.Errorf("expected course 'Green Valley', got '%s'", event.CourseName)
	}
	// Host should be accepted
	if len(event.Accepted) != 1 || event.Accepted[0] != p1 {
		t.Errorf("expected host in accepted list, got %v", event.Accepted)
	}
	// Other player should be pending (public event invites all)
	if len(event.Pending) != 1 || event.Pending[0] != p2 {
		t.Errorf("expected other player in pending list, got %v", event.Pending)
	}
}

func TestCreateEvent_Private(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	p3 := seedPlayer(t, "Cleo", "cleo@test.com", "password")
	c1 := seedCourse(t, "Green Valley")

	body := map[string]interface{}{
		"course_id":       c1,
		"date":            "2025-08-01",
		"tee_time":        "10:00",
		"open_spots":      3,
		"number_of_holes": "18",
		"private":         true,
		"host_id":         p1,
		"invitees":        []int64{p2},
	}

	rr := doRequest(t, "POST", "/api/v1/event", body, testHandler.CreateEvent)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var event model.EventResponse
	json.NewDecoder(rr.Body).Decode(&event)
	// Host accepted, p2 pending, p3 not invited
	if len(event.Accepted) != 1 {
		t.Errorf("expected 1 accepted, got %d", len(event.Accepted))
	}
	if len(event.Pending) != 1 || event.Pending[0] != p2 {
		t.Errorf("expected p2 in pending, got %v", event.Pending)
	}
	_ = p3 // p3 should not be invited
}

func TestCreateEvent_MissingFields(t *testing.T) {
	cleanDB(t)

	body := map[string]interface{}{
		"date": "2025-08-01",
	}

	rr := doRequest(t, "POST", "/api/v1/event", body, testHandler.CreateEvent)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestDeleteEvent(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, p1, 3, false)
	seedPlayerEvent(t, p1, eid, 1)

	rr := doRequestWithChiCtx(t, "DELETE", fmt.Sprintf("/api/v1/event/%d", eid), nil, testHandler.DeleteEvent, map[string]string{"id": fmt.Sprintf("%d", eid)})

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify event is gone
	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM events WHERE id = $1", eid).Scan(&count)
	if count != 0 {
		t.Error("event should be deleted")
	}

	// Verify player_events cascaded
	testDB.QueryRow("SELECT COUNT(*) FROM player_events WHERE event_id = $1", eid).Scan(&count)
	if count != 0 {
		t.Error("player_events should be cascade deleted")
	}
}

// ===================== EVENTS BY PLAYER =====================

func TestListEventsByPlayer(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	e1 := seedEvent(t, c1, p1, 3, false)
	e2 := seedEvent(t, c1, p2, 3, false)
	seedPlayerEvent(t, p1, e1, 1)
	seedPlayerEvent(t, p1, e2, 0)
	seedPlayerEvent(t, p2, e2, 1)

	// Use chi context for player_id URL param
	rr := doRequestWithChiCtx(t, "GET", fmt.Sprintf("/api/v1/players/%d/events", p1), nil, testHandler.ListEvents, map[string]string{"player_id": fmt.Sprintf("%d", p1)})

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var events []model.EventResponse
	json.NewDecoder(rr.Body).Decode(&events)
	if len(events) != 2 {
		t.Errorf("expected 2 events for player, got %d", len(events))
	}
}

// ===================== FRIENDSHIPS =====================

func TestCreateFriendship(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")

	body := map[string]int64{
		"follower_id": p1,
		"followee_id": p2,
	}

	rr := doRequest(t, "POST", "/api/v1/friendship", body, testHandler.CreateFriendship)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var friendship model.FriendshipResponse
	json.NewDecoder(rr.Body).Decode(&friendship)
	if friendship.Followee.Name != "Bob" {
		t.Errorf("expected followee name 'Bob', got '%s'", friendship.Followee.Name)
	}
}

func TestCreateFriendship_Idempotent(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	seedFriendship(t, p1, p2)

	body := map[string]int64{
		"follower_id": p1,
		"followee_id": p2,
	}

	rr := doRequest(t, "POST", "/api/v1/friendship", body, testHandler.CreateFriendship)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 (idempotent), got %d", rr.Code)
	}

	// Should still be only one friendship
	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM friendships WHERE follower_id = $1 AND followee_id = $2", p1, p2).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 friendship, got %d", count)
	}
}

func TestDeleteFriendship(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	seedFriendship(t, p1, p2)

	body := map[string]int64{
		"follower_id": p1,
		"followee_id": p2,
	}

	rr := doRequest(t, "DELETE", "/api/v1/friendship", body, testHandler.DeleteFriendship)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM friendships WHERE follower_id = $1 AND followee_id = $2", p1, p2).Scan(&count)
	if count != 0 {
		t.Error("friendship should be deleted")
	}
}

// ===================== PLAYER EVENTS =====================

func TestUpdatePlayerEvent_Accept(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, p1, 3, false)
	seedPlayerEvent(t, p1, eid, 0) // pending

	body := map[string]interface{}{
		"player_id":     p1,
		"event_id":      eid,
		"invite_status": "accepted",
	}

	rr := doRequest(t, "PATCH", "/api/v1/player-event", body, testHandler.UpdatePlayerEvent)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var pe model.PlayerEventResponse
	json.NewDecoder(rr.Body).Decode(&pe)
	if pe.InviteStatus != "accepted" {
		t.Errorf("expected status 'accepted', got '%s'", pe.InviteStatus)
	}
}

func TestUpdatePlayerEvent_CascadeClose(t *testing.T) {
	cleanDB(t)
	host := seedPlayer(t, "Host", "host@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	p3 := seedPlayer(t, "Cleo", "cleo@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, host, 2, false) // 2 spots total

	seedPlayerEvent(t, host, eid, 1) // host accepted (1 of 2 spots)
	seedPlayerEvent(t, p2, eid, 0)   // pending
	seedPlayerEvent(t, p3, eid, 0)   // pending

	// p2 accepts — fills last spot
	body := map[string]interface{}{
		"player_id":     p2,
		"event_id":      eid,
		"invite_status": "accepted",
	}

	rr := doRequest(t, "PATCH", "/api/v1/player-event", body, testHandler.UpdatePlayerEvent)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// p3 should now be closed (was pending, spots are full)
	var status int
	testDB.QueryRow("SELECT invite_status FROM player_events WHERE player_id = $1 AND event_id = $2", p3, eid).Scan(&status)
	if status != 3 { // closed
		t.Errorf("expected p3 status to be 3 (closed), got %d", status)
	}
}

func TestUpdatePlayerEvent_CascadeReopen(t *testing.T) {
	cleanDB(t)
	host := seedPlayer(t, "Host", "host@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	p3 := seedPlayer(t, "Cleo", "cleo@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, host, 2, false) // 2 spots total

	seedPlayerEvent(t, host, eid, 1) // host accepted
	seedPlayerEvent(t, p2, eid, 1)   // accepted (now full)
	seedPlayerEvent(t, p3, eid, 3)   // closed

	// p2 declines — frees a spot
	body := map[string]interface{}{
		"player_id":     p2,
		"event_id":      eid,
		"invite_status": "declined",
	}

	rr := doRequest(t, "PATCH", "/api/v1/player-event", body, testHandler.UpdatePlayerEvent)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// p3 should now be reopened to pending
	var status int
	testDB.QueryRow("SELECT invite_status FROM player_events WHERE player_id = $1 AND event_id = $2", p3, eid).Scan(&status)
	if status != 0 { // pending
		t.Errorf("expected p3 status to be 0 (pending), got %d", status)
	}
}

func TestUpdatePlayerEvent_Decline(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, p1, 3, false)
	seedPlayerEvent(t, p1, eid, 0) // pending

	body := map[string]interface{}{
		"player_id":     p1,
		"event_id":      eid,
		"invite_status": "declined",
	}

	rr := doRequest(t, "PATCH", "/api/v1/player-event", body, testHandler.UpdatePlayerEvent)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var pe model.PlayerEventResponse
	json.NewDecoder(rr.Body).Decode(&pe)
	if pe.InviteStatus != "declined" {
		t.Errorf("expected status 'declined', got '%s'", pe.InviteStatus)
	}
}

// ===================== PLAYER WITH DETAILS =====================

func TestPlayerResponse_IncludesFriendsAndEvents(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, p1, 3, false)

	seedFriendship(t, p1, p2)
	seedPlayerEvent(t, p1, eid, 1) // accepted

	rr := doRequest(t, "GET", "/api/v1/players", nil, testHandler.ListPlayers)

	var players []model.PlayerResponse
	json.NewDecoder(rr.Body).Decode(&players)

	// Find Amy
	var amy model.PlayerResponse
	for _, p := range players {
		if p.Name == "Amy" {
			amy = p
			break
		}
	}

	if len(amy.Friends) != 1 || amy.Friends[0] != p2 {
		t.Errorf("expected Amy to have Bob as friend, got %v", amy.Friends)
	}
	if len(amy.Events) != 1 || amy.Events[0] != eid {
		t.Errorf("expected Amy to have event %d, got %v", eid, amy.Events)
	}
}

// ===================== SESSION WITH DETAILS =====================

func TestSessionResponse_IncludesFriendsAndEvents(t *testing.T) {
	cleanDB(t)
	p1 := seedPlayer(t, "Amy", "amy@test.com", "password")
	p2 := seedPlayer(t, "Bob", "bob@test.com", "password")
	c1 := seedCourse(t, "Green Valley")
	eid := seedEvent(t, c1, p1, 3, false)

	seedFriendship(t, p1, p2)
	seedPlayerEvent(t, p1, eid, 1)

	body := map[string]string{
		"email":    "amy@test.com",
		"password": "password",
	}

	rr := doRequest(t, "POST", "/api/v1/sessions", body, testHandler.CreateSession)

	var login model.LoginResponse
	json.NewDecoder(rr.Body).Decode(&login)

	if len(login.Friends) != 1 {
		t.Errorf("expected 1 friend in login response, got %d", len(login.Friends))
	}
	if len(login.Events) != 1 {
		t.Errorf("expected 1 event in login response, got %d", len(login.Events))
	}
}
